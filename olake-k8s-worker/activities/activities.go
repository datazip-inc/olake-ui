package activities

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"crypto/sha256"
	"olake-k8s-worker/database"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/utils"
)

// DiscoverCatalogActivity - SIMPLIFIED VERSION
func DiscoverCatalogActivity(ctx context.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
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

	// Create and run the Job (this will write config files to PV)
	job, err := jobManager.CreateJobWithPV(ctx, jobSpec, configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	// Wait for completion
	result, err := jobManager.WaitForJobCompletion(ctx, job.Name, 5*time.Minute)
	if err != nil {
		// Cleanup only needs to delete the Job (no ConfigMap)
		jobManager.CleanupJob(ctx, job.Name)
		return nil, fmt.Errorf("discover job failed: %v", err)
	}

	// ✅ ADD BACK: Read the streams.json file from the PV (like Docker does)
	workflowDir := params.WorkflowID // Use original WorkflowID for directory
	streamsPath := filepath.Join("/data/olake-jobs", workflowDir, "streams.json")
	catalogData, err := utils.ParseJSONFile(streamsPath)
	if err != nil {
		logger.Errorf("Failed to read streams.json: %v", err)
		return result, nil // Return the parsed pod logs as fallback
	}

	// Cleanup resources
	if cleanupErr := jobManager.CleanupJob(ctx, job.Name); cleanupErr != nil {
		logger.Warnf("Failed to cleanup job: %v", cleanupErr)
	}

	return catalogData, nil // ✅ Return the actual catalog data from streams.json
}

// TestConnectionActivity runs the check command using Kubernetes Jobs
func TestConnectionActivity(ctx context.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
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

	// Log the command that will be executed
	commandStr := strings.Join(jobSpec.Args, " ")
	logger.Infof("Creating test connection job with image: %s", jobSpec.Image)
	logger.Infof("Pod command: %s", commandStr)

	// Create and run the Job
	job, err := jobManager.CreateJobWithPV(ctx, jobSpec, configs)
	if err != nil {
		logger.Errorf("Failed to create job: %v", err)
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	activityLogger.Info("Created Kubernetes Job", "jobName", job.Name)
	logger.Infof("Successfully created Kubernetes Job: %s", job.Name)

	// Wait for completion
	result, err := jobManager.WaitForJobCompletion(ctx, job.Name, 5*time.Minute)
	if err != nil {
		activityLogger.Error("Job failed", "error", err)
		logger.Errorf("Test connection job failed: %v", err)
		// Cleanup even on failure
		jobManager.CleanupJob(ctx, job.Name)
		return nil, fmt.Errorf("test connection job failed: %v", err)
	}

	// Add delay to ensure logs are fully available
	logger.Debugf("Waiting 5 seconds before cleanup to ensure logs are fully retrieved")
	time.Sleep(5 * time.Second)

	// Cleanup resources
	if cleanupErr := jobManager.CleanupJob(ctx, job.Name); cleanupErr != nil {
		activityLogger.Warn("Failed to cleanup resources", "error", cleanupErr)
		logger.Warnf("Failed to cleanup resources: %v", cleanupErr)
	}

	activityLogger.Info("Connection test completed successfully")
	logger.Info("Connection test completed successfully")
	return result, nil
}

// SyncActivity runs the sync command using Kubernetes Jobs
func SyncActivity(ctx context.Context, params *shared.SyncParams) (map[string]interface{}, error) {
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

	// Log the command that will be executed
	commandStr := strings.Join(jobSpec.Args, " ")
	logger.Infof("Creating sync job with image: %s", jobSpec.Image)
	logger.Infof("Pod command: %s", commandStr)

	// Create and run the Job
	job, err := jobManager.CreateJobWithPV(ctx, jobSpec, configs)
	if err != nil {
		logger.Errorf("Failed to create job: %v", err)
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	activityLogger.Info("Created Kubernetes Job", "jobName", job.Name)
	logger.Infof("Successfully created Kubernetes Job: %s", job.Name)

	// Wait for completion (longer timeout for sync operations)
	result, err := jobManager.WaitForJobCompletion(ctx, job.Name, 15*time.Minute)
	if err != nil {
		activityLogger.Error("Job failed", "error", err)
		logger.Errorf("Sync job failed: %v", err)
		// Cleanup even on failure
		jobManager.CleanupJob(ctx, job.Name)
		return nil, fmt.Errorf("sync job failed: %v", err)
	}

	// ✅ ADD: Read updated state from PV (like Docker does)
	workflowDir := fmt.Sprintf("%x", sha256.Sum256([]byte(params.WorkflowID)))
	statePath := filepath.Join("/data/olake-jobs", workflowDir, "state.json")
	updatedState, err := utils.ParseJSONFile(statePath)
	if err != nil {
		logger.Warnf("Failed to read updated state file: %v", err)
		updatedState = result // Fallback to pod logs
	}

	// ✅ ADD: Update job state in database with updated state
	if updateErr := UpdateJobState(params.JobID, updatedState); updateErr != nil {
		logger.Warnf("Failed to update job state for jobID %d: %v", params.JobID, updateErr)
	}

	// Cleanup resources
	if cleanupErr := jobManager.CleanupJob(ctx, job.Name); cleanupErr != nil {
		activityLogger.Warn("Failed to cleanup resources", "error", cleanupErr)
		logger.Warnf("Failed to cleanup resources: %v", cleanupErr)
	}

	activityLogger.Info("Data sync completed successfully")
	logger.Info("Data sync completed successfully")
	return result, nil
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
