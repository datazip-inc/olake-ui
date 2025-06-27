package activities

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"olake-k8s-worker/database"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/utils"
)

// DiscoverCatalogActivity - Convert to Pod
func DiscoverCatalogActivity(ctx context.Context, params shared.ActivityParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Info("Starting K8s discover catalog activity")

	// Create K8s Job manager
	jobManager, err := NewK8sJobManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s job manager: %v", err)
	}

	// Prepare config files (but don't create ConfigMap)
	configs := []shared.JobConfig{
		{Name: "config.json", Data: params.Config},
	}

	// Create Job specification
	jobSpec := &JobSpec{
		Name:               utils.SanitizeK8sName(params.WorkflowID),
		OriginalWorkflowID: params.WorkflowID,
		Image:              jobManager.GetDockerImageName(params.SourceType, params.Version),
		Command:            []string{},
		Args:               []string{string(shared.Discover), "--config", "/mnt/config/config.json"},
		Operation:          shared.Discover,
	}

	// Create Pod instead of Job
	pod, err := jobManager.CreatePod(ctx, jobSpec, configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// DISABLED: Always cleanup pod when done
	// defer func() {
	// 	if err := jobManager.CleanupPod(ctx, pod.Name); err != nil {
	// 		logger.Errorf("Failed to cleanup pod %s: %v", pod.Name, err)
	// 	}
	// }()

	// Wait for pod completion
	return jobManager.WaitForPodCompletion(ctx, pod.Name, 5*time.Minute)
}

// TestConnectionActivity - Convert to Pod
func TestConnectionActivity(ctx context.Context, params shared.ActivityParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Info("Starting K8s test connection activity",
		"sourceType", params.SourceType,
		"workflowID", params.WorkflowID)

	logger.Infof("Starting test connection activity for sourceType: %s, version: %s, workflowID: %s",
		params.SourceType, params.Version, params.WorkflowID)

	// Create K8s Job manager
	jobManager, err := NewK8sJobManager()
	if err != nil {
		logger.Errorf("Failed to create K8s job manager: %v", err)
		return nil, fmt.Errorf("failed to create K8s job manager: %v", err)
	}

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Job for connection test")

	// Create ConfigMap with source configuration
	configs := []shared.JobConfig{
		{Name: "config.json", Data: params.Config},
	}

	// Create Job specification
	jobSpec := &JobSpec{
		Name:               utils.SanitizeK8sName(params.WorkflowID),
		OriginalWorkflowID: params.WorkflowID,
		Image:              jobManager.GetDockerImageName(params.SourceType, params.Version),
		Command:            []string{},
		Args: []string{
			string(shared.Check),
			fmt.Sprintf("--%s", params.Flag),
			"/mnt/config/config.json",
		},
		Operation: shared.Check,
	}

	// Create Pod instead of Job
	pod, err := jobManager.CreatePod(ctx, jobSpec, configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// DISABLED: Always cleanup pod when done
	// defer func() {
	// 	if err := jobManager.CleanupPod(ctx, pod.Name); err != nil {
	// 		logger.Errorf("Failed to cleanup pod %s: %v", pod.Name, err)
	// 	}
	// }()

	// Wait for pod completion
	return jobManager.WaitForPodCompletion(ctx, pod.Name, 5*time.Minute)
}

// SyncActivity - Convert to Pod
func SyncActivity(ctx context.Context, params shared.SyncParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Info("Starting K8s sync activity",
		"jobId", params.JobID,
		"workflowID", params.WorkflowID)

	logger.Infof("Starting sync activity for jobID: %d, workflowID: %s", params.JobID, params.WorkflowID)

	// Create K8s Job manager
	jobManager, err := NewK8sJobManager()
	if err != nil {
		logger.Errorf("Failed to create K8s job manager: %v", err)
		return nil, fmt.Errorf("failed to create K8s job manager: %v", err)
	}

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Job for data sync")

	// Get job details from database (we'll need to adapt this)
	jobData, err := GetJobData(params.JobID)
	if err != nil {
		logger.Errorf("Failed to get job data for jobID %d: %v", params.JobID, err)
		return nil, fmt.Errorf("failed to get job data: %v", err)
	}

	// ✅ ADD: Validate and fix empty/null state
	stateData := jobData.State
	if stateData == "" || stateData == "null" || stateData == "NULL" {
		stateData = "{}"
		logger.Infof("Job %d has empty/null state, defaulting to: {}", params.JobID)
	}

	// Create ConfigMap with all necessary configuration files
	configs := []shared.JobConfig{
		{Name: "config.json", Data: jobData.SourceConfig},
		{Name: "streams.json", Data: jobData.StreamsConfig},
		{Name: "writer.json", Data: jobData.DestConfig},
		{Name: "state.json", Data: stateData},
	}

	// Create Job specification for sync
	jobSpec := &JobSpec{
		Name:               utils.SanitizeK8sName(params.WorkflowID),
		OriginalWorkflowID: params.WorkflowID,
		Image:              jobManager.GetDockerImageName(jobData.SourceType, jobData.SourceVersion),
		Command:            []string{},
		Args: []string{
			string(shared.Sync),
			"--config", "/mnt/config/config.json",
			"--catalog", "/mnt/config/streams.json",
			"--destination", "/mnt/config/writer.json",
			"--state", "/mnt/config/state.json",
		},
		Operation: shared.Sync,
	}

	// Create Pod instead of Job
	pod, err := jobManager.CreatePod(ctx, jobSpec, configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// DISABLED: Always cleanup pod when done
	// defer func() {
	// 	if err := jobManager.CleanupPod(ctx, pod.Name); err != nil {
	// 		logger.Errorf("Failed to cleanup pod %s: %v", pod.Name, err)
	// 	}
	// }()

	// Wait for pod completion
	return jobManager.WaitForPodCompletion(ctx, pod.Name, 15*time.Minute)
}

// GetJobData fetches job configuration from database
func GetJobData(jobID int) (*JobData, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	jobData, err := db.GetJobData(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	// Convert database JobData to activity JobData
	return &JobData{
		SourceType:    jobData.SourceType,
		SourceVersion: jobData.SourceVersion,
		SourceConfig:  jobData.SourceConfig,
		DestConfig:    jobData.DestConfig,
		StreamsConfig: jobData.StreamsConfig,
		State:         jobData.State,
	}, nil
}

// UpdateJobState updates job state in database
func UpdateJobState(jobID int, state map[string]interface{}) error {
	db, err := database.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	return db.UpdateJobState(jobID, state)
}

// ParseJobOutput extracts JSON from Kubernetes job logs
func ParseJobOutput(output string) (map[string]interface{}, error) {
	// Use the flexible parser that can handle different output formats
	return utils.ParseJobOutput(output)
}

type JobData struct {
	SourceType    string
	SourceVersion string
	SourceConfig  string
	DestConfig    string
	StreamsConfig string
	State         string
}
