package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/activity"

	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/config/helpers"
	"olake-ui/olake-workers/k8s/database/service"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/pods"
	"olake-ui/olake-workers/k8s/shared"
	"olake-ui/olake-workers/k8s/utils/filesystem"
)

// Activities holds the dependencies for activity functions
type Activities struct {
	jobService service.JobDataService
	podManager *pods.K8sPodManager
	config     *config.Config
}

// NewActivities creates a new Activities instance with injected dependencies
func NewActivities(jobService service.JobDataService, podManager *pods.K8sPodManager, cfg *config.Config) *Activities {
	return &Activities{
		jobService: jobService,
		podManager: podManager,
		config:     cfg,
	}
}

// DiscoverCatalogActivity discovers data source catalog using Kubernetes Pod
func (a *Activities) DiscoverCatalogActivity(ctx context.Context, params shared.ActivityParams) (map[string]interface{}, error) {
	activityLogger := activity.GetLogger(ctx)
	activityLogger.Debug("Starting K8s discover catalog activity")

	// Use injected pod manager

	// Execute pod activity using common workflow
	request := pods.PodActivityRequest{
		WorkflowID: params.WorkflowID,
		JobID:      params.JobID,
		Operation:  shared.Discover,
		Image:      a.podManager.GetDockerImageName(params.SourceType, params.Version),
		Args:       []string{string(shared.Discover), "--config", "/mnt/config/config.json"},
		Configs: []shared.JobConfig{
			{Name: "config.json", Data: params.Config},
		},
		Timeout: helpers.GetActivityTimeout(a.config, "discover"),
	}

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

	// Execute pod activity using common workflow
	request := pods.PodActivityRequest{
		WorkflowID: params.WorkflowID,
		JobID:      params.JobID,
		Operation:  shared.Check,
		Image:      a.podManager.GetDockerImageName(params.SourceType, params.Version),
		Args: []string{
			string(shared.Check),
			fmt.Sprintf("--%s", params.Flag),
			"/mnt/config/config.json",
		},
		Configs: []shared.JobConfig{
			{Name: "config.json", Data: params.Config},
		},
		Timeout: helpers.GetActivityTimeout(a.config, "test"),
	}

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

	// Start state monitoring goroutine for incremental persistence
	go a.monitorState(ctx, params.JobID, params.WorkflowID)

	// Get job details from database using injected service
	jobData, err := a.jobService.GetJobData(params.JobID)
	if err != nil {
		logger.Errorf("Failed to get job data for jobID %d: %v", params.JobID, err)
		return nil, fmt.Errorf("failed to get job data: %v", err)
	}

	// Validate and fix empty/null state
	stateData := jobData.State
	if stateData == "" || stateData == "null" || stateData == "NULL" {
		stateData = "{}"
		logger.Infof("Job %d has empty/null state, defaulting to: {}", params.JobID)
	}

	// Execute pod activity using common workflow
	request := pods.PodActivityRequest{
		WorkflowID: params.WorkflowID,
		JobID:      params.JobID,
		Operation:  shared.Sync,
		Image:      a.podManager.GetDockerImageName(jobData.SourceType, jobData.SourceVersion),
		Args: []string{
			string(shared.Sync),
			"--config", "/mnt/config/config.json",
			"--catalog", "/mnt/config/streams.json",
			"--destination", "/mnt/config/writer.json",
			"--state", "/mnt/config/state.json",
		},
		Configs: []shared.JobConfig{
			{Name: "config.json", Data: jobData.SourceConfig},
			{Name: "streams.json", Data: jobData.StreamsConfig},
			{Name: "writer.json", Data: jobData.DestConfig},
			{Name: "state.json", Data: stateData},
		},
		Timeout: helpers.GetActivityTimeout(a.config, "sync"),
	}

	result, err := a.podManager.ExecutePodActivity(ctx, request)
	if err != nil {
		return nil, err
	}

	// Update job state similar to server implementation
	if stateJSON, err := json.Marshal(result); err == nil {
		if err := a.jobService.UpdateJobState(params.JobID, string(stateJSON), true); err != nil {
			logger.Errorf("Failed to update job state for jobID %d: %v", params.JobID, err)
			return nil, fmt.Errorf("failed to update job state: %v", err)
		}
		logger.Infof("Successfully updated job state for jobID %d", params.JobID)
	} else {
		logger.Warnf("Failed to marshal result for jobID %d: %v", params.JobID, err)
	}

	return result, nil
}

// monitorState monitors state.json file changes and persists incremental updates to PostgreSQL
func (a *Activities) monitorState(ctx context.Context, jobID int, workflowID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	var lastModTime time.Time
	
	for {
		select {
		case <-ticker.C:
			if err := a.checkpointState(jobID, workflowID, &lastModTime); err != nil {
				logger.Errorf("Failed to checkpoint state for job %d: %v", jobID, err)
			}
		case <-ctx.Done():
			// Final checkpoint on shutdown
			logger.Infof("Activity context cancelled, performing final state checkpoint for job %d", jobID)
			if err := a.checkpointState(jobID, workflowID, &lastModTime); err != nil {
				logger.Errorf("Final checkpoint failed for job %d: %v", jobID, err)
			}
			return
		}
	}
}

// checkpointState reads state.json and persists it to PostgreSQL if it has changed
func (a *Activities) checkpointState(jobID int, workflowID string, lastModTime *time.Time) error {
	// Create filesystem helper to get correct directory path
	fsHelper := filesystem.NewHelper()
	workflowDir := fsHelper.GetWorkflowDirectory(shared.Sync, workflowID)
	statePath := fsHelper.GetFilePath(workflowDir, "state.json")
	
	// Check if file changed using ModTime
	info, err := os.Stat(statePath)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	
	if !info.ModTime().After(*lastModTime) {
		return nil // No change
	}
	
	// Read state file
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}
	
	// Validate JSON
	var js json.RawMessage
	if err := json.Unmarshal(stateData, &js); err != nil {
		return fmt.Errorf("invalid JSON, skipping checkpoint: %w", err)
	}
	
	// Check for empty/truncated file
	if len(stateData) < 10 {
		return fmt.Errorf("state file too small (%d bytes), skipping", len(stateData))
	}
	
	// Persist to PostgreSQL using existing service
	if err := a.jobService.UpdateJobState(jobID, string(stateData), true); err != nil {
		return fmt.Errorf("db update failed: %w", err)
	}
	
	*lastModTime = info.ModTime()
	logger.Debugf("State checkpoint completed for job %d", jobID)
	return nil
}
