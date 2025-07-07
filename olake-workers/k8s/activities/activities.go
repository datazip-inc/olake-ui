package activities

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/config/helpers"
	"olake-ui/olake-workers/k8s/database/service"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/pods"
	"olake-ui/olake-workers/k8s/shared"
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
	activityLogger.Info("Starting K8s discover catalog activity")

	// Use injected pod manager

	// Execute pod activity using common workflow
	request := pods.PodActivityRequest{
		WorkflowID: params.WorkflowID,
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
	activityLogger.Info("Starting K8s test connection activity",
		"sourceType", params.SourceType,
		"workflowID", params.WorkflowID)

	logger.Infof("Starting test connection activity for sourceType: %s, version: %s, workflowID: %s",
		params.SourceType, params.Version, params.WorkflowID)

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Pod for connection test")

	// Execute pod activity using common workflow
	request := pods.PodActivityRequest{
		WorkflowID: params.WorkflowID,
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
	activityLogger.Info("Starting K8s sync activity",
		"jobId", params.JobID,
		"workflowID", params.WorkflowID)

	logger.Infof("Starting sync activity for jobID: %d, workflowID: %s", params.JobID, params.WorkflowID)

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Creating Kubernetes Pod for data sync")

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

	return a.podManager.ExecutePodActivity(ctx, request)
}
