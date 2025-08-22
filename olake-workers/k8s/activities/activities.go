package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/activity"

	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/database"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/pods"
	"olake-ui/olake-workers/k8s/shared"
	"olake-ui/olake-workers/k8s/utils/filesystem"
	"olake-ui/olake-workers/k8s/utils/helpers"
)

// Activities holds the dependencies for activity functions
type Activities struct {
	db         *database.DB
	podManager *pods.K8sPodManager
	config     *config.Config
}

// NewActivities creates a new Activities instance with injected dependencies
func NewActivities(db *database.DB, podManager *pods.K8sPodManager, cfg *config.Config) *Activities {
	return &Activities{
		db:         db,
		podManager: podManager,
		config:     cfg,
	}
}

// DiscoverCatalogActivity discovers data source catalog using Kubernetes Pod
func (a *Activities) DiscoverCatalogActivity(ctx context.Context, params shared.ActivityParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Debug("Starting K8s discover catalog activity")

	// Transform Temporal activity parameters into Kubernetes pod execution request
	// Maps connector type/version to container image, mounts config as files, sets operation-specific timeout
	imageName, err := a.podManager.GetDockerImageName(params.SourceType, params.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker image name: %v", err)
	}

	request := pods.PodActivityRequest{
		WorkflowID:    params.WorkflowID,
		JobID:         params.JobID,
		Operation:     shared.Discover,
		ConnectorType: params.SourceType,
		Image:         imageName,
		Args:          []string{string(shared.Discover), "--config", "/mnt/config/config.json"},
		Configs: []shared.JobConfig{
			{Name: "config.json", Data: params.Config},
		},
		Timeout: helpers.GetActivityTimeout("discover"),
	}

	// Execute discover operation by creating K8s pod, wait for completion, retrieve results from streams.json file
	return a.podManager.ExecutePodActivity(ctx, request)
}

// TestConnectionActivity tests data source connection using Kubernetes Pod
func (a *Activities) TestConnectionActivity(ctx context.Context, params shared.ActivityParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Debug("Starting K8s test connection activity",
		"sourceType", params.SourceType,
		"version", params.Version,
		"workflowID", params.WorkflowID)

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Pod for connection test")

	// Transform Temporal activity parameters into Kubernetes pod execution request
	// Maps connector type/version to container image, includes flag parameter, mounts config as files
	imageName, err := a.podManager.GetDockerImageName(params.SourceType, params.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker image name: %v", err)
	}

	request := pods.PodActivityRequest{
		WorkflowID:    params.WorkflowID,
		JobID:         params.JobID,
		Operation:     shared.Check,
		ConnectorType: params.SourceType,
		Image:         imageName,
		Args: []string{
			string(shared.Check),
			fmt.Sprintf("--%s", params.Flag),
			"/mnt/config/config.json",
		},
		Configs: []shared.JobConfig{
			{Name: "config.json", Data: params.Config},
		},
		Timeout: helpers.GetActivityTimeout("test"),
	}

	// Execute check operation by creating K8s pod, wait for completion, retrieve results from pod logs
	return a.podManager.ExecutePodActivity(ctx, request)
}

// SyncActivity syncs data using Kubernetes Pod
func (a *Activities) SyncActivity(ctx context.Context, params shared.SyncParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Debug("Starting K8s sync activity",
		"jobId", params.JobID,
		"workflowID", params.WorkflowID)

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Pod for data sync")

	// Retrieve job configuration from database to get all required sync parameters
	jobData, err := a.db.GetJobData(ctx, params.JobID)
	if err != nil {
		logger.Errorf("Failed to get job data for jobID %d: %v", params.JobID, err)
		return nil, fmt.Errorf("failed to get job data: %v", err)
	}

	// Validate and fix empty/null state
	stateData := jobData["state"].(string)
	if stateData == "" || stateData == "null" || stateData == "NULL" {
		stateData = "{}"
		logger.Infof("Job %d has empty/null state, defaulting to: {}", params.JobID)
	}

	// Transform job data and Temporal activity parameters into Kubernetes pod execution request
	// Maps all sync configuration files (config, catalog, destination, state) as mounted files
	imageName, err := a.podManager.GetDockerImageName(jobData["source_type"].(string), jobData["source_version"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to get docker image name: %v", err)
	}

	request := pods.PodActivityRequest{
		WorkflowID:    params.WorkflowID,
		JobID:         params.JobID,
		Operation:     shared.Sync,
		ConnectorType: jobData["source_type"].(string),
		Image:         imageName,
		Args: []string{
			string(shared.Sync),
			"--config", "/mnt/config/config.json",
			"--catalog", "/mnt/config/streams.json",
			"--destination", "/mnt/config/writer.json",
			"--state", "/mnt/config/state.json",
		},
		Configs: []shared.JobConfig{
			{Name: "config.json", Data: jobData["source_config"].(string)},
			{Name: "streams.json", Data: jobData["streams_config"].(string)},
			{Name: "writer.json", Data: jobData["dest_config"].(string)},
			{Name: "state.json", Data: stateData},
		},
		Timeout: helpers.GetActivityTimeout("sync"),
	}

	// Execute sync operation by creating K8s pod, wait for completion, retrieve results from state.json file
	result, err := a.podManager.ExecutePodActivity(ctx, request)
	if err != nil {
		logger.Warnf("Activity failed for job %d: %v. Attempting final state save", params.JobID, err)

		// Attempt to read final state from shared filesystem even on failure for data recovery
		fsHelper := filesystem.NewHelper()
		if stateData, readErr := fsHelper.ReadAndValidateStateFile(params.WorkflowID); readErr != nil {
			// Log if reading or validation fails, but don't block the process
			// This covers file not existing or containing invalid JSON
			logger.Warnf("Failed to read/validate final state on error: %v", readErr)
		} else {
			// If the state file is valid, attempt to save it
			if updateErr := a.db.UpdateJobState(ctx, params.JobID, string(stateData), false); updateErr != nil {
				logger.Errorf("Failed to save final state on error for job %d: %v", params.JobID, updateErr)
			} else {
				logger.Infof("Saved final state on failure for job %d", params.JobID)
			}
		}

		return nil, err
	}

	// Persist final sync state back to database for job tracking and resume capabilities
	if stateJSON, err := json.Marshal(result); err != nil {
		logger.Warnf("Failed to marshal result for jobID %d: %v", params.JobID, err)
	} else {
		if err := a.db.UpdateJobState(ctx, params.JobID, string(stateJSON), true); err != nil {
			logger.Errorf("Failed to update job state for jobID %d: %v", params.JobID, err)
			return nil, fmt.Errorf("failed to update job state: %v", err)
		}
		logger.Infof("Successfully updated job state for jobID %d", params.JobID)
	}

	return result, nil
}
