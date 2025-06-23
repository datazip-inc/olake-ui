package activities

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
)

// DiscoverCatalogActivity runs the discover command using Kubernetes Jobs
func DiscoverCatalogActivity(ctx context.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Info("Starting K8s discover catalog activity",
		"sourceType", params.SourceType,
		"workflowID", params.WorkflowID)

	logger.Infof("Starting discover catalog activity for sourceType: %s, version: %s, workflowID: %s",
		params.SourceType, params.Version, params.WorkflowID)

	// Create K8s Job manager
	jobManager, err := NewK8sJobManager()
	if err != nil {
		logger.Errorf("Failed to create K8s job manager: %v", err)
		return nil, fmt.Errorf("failed to create K8s job manager: %v", err)
	}

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Job for catalog discovery")

	// Create ConfigMap with source configuration
	configMapName := fmt.Sprintf("discover-config-%s", params.WorkflowID)
	configs := []shared.JobConfig{
		{Name: "config.json", Data: params.Config},
	}

	logger.Debugf("Creating ConfigMap %s with configuration files", configMapName)
	_, err = jobManager.CreateConfigMap(ctx, configMapName, configs)
	if err != nil {
		logger.Errorf("Failed to create ConfigMap: %v", err)
		return nil, fmt.Errorf("failed to create configmap: %v", err)
	}

	// Create Job specification
	jobSpec := &JobSpec{
		Name:          fmt.Sprintf("discover-%s", params.WorkflowID),
		Image:         jobManager.GetDockerImageName(params.SourceType, params.Version),
		Command:       []string{string(shared.Discover)},
		Args:          []string{"--config", "/mnt/config/config.json"},
		ConfigMapName: configMapName,
		Operation:     shared.Discover,
	}

	logger.Infof("Creating discover job with image: %s", jobSpec.Image)

	// Create and run the Job
	job, err := jobManager.CreateJob(ctx, jobSpec)
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
		logger.Errorf("Discover job failed: %v", err)
		// Cleanup even on failure
		jobManager.CleanupJob(ctx, job.Name, configMapName)
		return nil, fmt.Errorf("discover job failed: %v", err)
	}

	// Cleanup resources
	if cleanupErr := jobManager.CleanupJob(ctx, job.Name, configMapName); cleanupErr != nil {
		activityLogger.Warn("Failed to cleanup resources", "error", cleanupErr)
		logger.Warnf("Failed to cleanup resources: %v", cleanupErr)
	}

	activityLogger.Info("Catalog discovery completed successfully")
	logger.Info("Catalog discovery completed successfully")
	return result, nil
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
	configMapName := fmt.Sprintf("test-config-%s", params.WorkflowID)
	configs := []shared.JobConfig{
		{Name: "config.json", Data: params.Config},
	}

	logger.Debugf("Creating ConfigMap %s for connection test", configMapName)
	_, err = jobManager.CreateConfigMap(ctx, configMapName, configs)
	if err != nil {
		logger.Errorf("Failed to create ConfigMap: %v", err)
		return nil, fmt.Errorf("failed to create configmap: %v", err)
	}

	// Create Job specification
	jobSpec := &JobSpec{
		Name:          fmt.Sprintf("test-%s", params.WorkflowID),
		Image:         jobManager.GetDockerImageName(params.SourceType, params.Version),
		Command:       []string{string(shared.Check)},
		Args:          []string{fmt.Sprintf("--%s", params.Flag), "/mnt/config/config.json"},
		ConfigMapName: configMapName,
		Operation:     shared.Check,
	}

	logger.Infof("Creating test connection job with image: %s", jobSpec.Image)

	// Create and run the Job
	job, err := jobManager.CreateJob(ctx, jobSpec)
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
		jobManager.CleanupJob(ctx, job.Name, configMapName)
		return nil, fmt.Errorf("test connection job failed: %v", err)
	}

	// Cleanup resources
	if cleanupErr := jobManager.CleanupJob(ctx, job.Name, configMapName); cleanupErr != nil {
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

	// Create ConfigMap with all necessary configuration files
	configMapName := fmt.Sprintf("sync-config-%s", params.WorkflowID)
	configs := []shared.JobConfig{
		{Name: "config.json", Data: jobData.SourceConfig},
		{Name: "streams.json", Data: jobData.StreamsConfig},
		{Name: "writer.json", Data: jobData.DestConfig},
		{Name: "state.json", Data: jobData.State},
	}

	logger.Debugf("Creating ConfigMap %s with all sync configuration files", configMapName)
	_, err = jobManager.CreateConfigMap(ctx, configMapName, configs)
	if err != nil {
		logger.Errorf("Failed to create ConfigMap: %v", err)
		return nil, fmt.Errorf("failed to create configmap: %v", err)
	}

	// Create Job specification for sync
	jobSpec := &JobSpec{
		Name:    fmt.Sprintf("sync-%s", params.WorkflowID),
		Image:   jobManager.GetDockerImageName(jobData.SourceType, jobData.SourceVersion),
		Command: []string{string(shared.Sync)},
		Args: []string{
			"--config", "/mnt/config/config.json",
			"--catalog", "/mnt/config/streams.json",
			"--destination", "/mnt/config/writer.json",
			"--state", "/mnt/config/state.json",
		},
		ConfigMapName: configMapName,
		Operation:     shared.Sync,
	}

	logger.Infof("Creating sync job with image: %s", jobSpec.Image)

	// Create and run the Job
	job, err := jobManager.CreateJob(ctx, jobSpec)
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
		jobManager.CleanupJob(ctx, job.Name, configMapName)
		return nil, fmt.Errorf("sync job failed: %v", err)
	}

	// Update job state in database
	if updateErr := UpdateJobState(params.JobID, result); updateErr != nil {
		activityLogger.Warn("Failed to update job state", "error", updateErr)
		logger.Warnf("Failed to update job state for jobID %d: %v", params.JobID, updateErr)
	}

	// Cleanup resources
	if cleanupErr := jobManager.CleanupJob(ctx, job.Name, configMapName); cleanupErr != nil {
		activityLogger.Warn("Failed to cleanup resources", "error", cleanupErr)
		logger.Warnf("Failed to cleanup resources: %v", cleanupErr)
	}

	activityLogger.Info("Data sync completed successfully")
	logger.Info("Data sync completed successfully")
	return result, nil
}

// Helper functions that will need to be implemented
func GetJobData(jobID int) (*JobData, error) {
	// TODO: Implement database access to get job configuration
	// This will need to connect to the same database as the server
	logger.Warnf("GetJobData not implemented for jobID: %d", jobID)
	return nil, fmt.Errorf("not implemented")
}

func UpdateJobState(jobID int, state map[string]interface{}) error {
	// TODO: Implement database update for job state
	logger.Warnf("UpdateJobState not implemented for jobID: %d", jobID)
	return fmt.Errorf("not implemented")
}

func ParseJobOutput(output string) (map[string]interface{}, error) {
	// TODO: Implement log parsing similar to Docker implementation
	// Extract JSON from container logs
	logger.Warn("ParseJobOutput not implemented")
	return nil, fmt.Errorf("not implemented")
}

type JobData struct {
	SourceType    string
	SourceVersion string
	SourceConfig  string
	DestConfig    string
	StreamsConfig string
	State         string
}
