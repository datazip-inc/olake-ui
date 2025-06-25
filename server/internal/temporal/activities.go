package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"go.temporal.io/sdk/activity"
)

// DiscoverCatalogActivity runs the discover command to get catalog data
func DiscoverCatalogActivity(ctx context.Context, params *ActivityParams) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity",
		"sourceType", params.SourceType,
		"workflowID", params.WorkflowID)

	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Running sync command")

	// Execute the sync operation
	result, err := runner.GetCatalog(
		ctx,
		params.SourceType,
		params.Version,
		params.Config,
		params.WorkflowID,
		params.StreamsConfig,
	)
	if err != nil {
		logger.Error("Sync command failed", "error", err)
		return result, fmt.Errorf("sync command failed: %v", err)
	}

	return result, nil
}

// // GetSpecActivity runs the spec command to get connector specifications
// func GetSpecActivity(ctx context.Context, params ActivityParams) (map[string]interface{}, error) {
// 	params.Command = docker.Spec
// 	return ExecuteDockerCommandActivity(ctx, params)
// }

// TestConnectionActivity runs the check command to test connection
func TestConnectionActivity(ctx context.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())
	resp, err := runner.TestConnection(ctx, params.Flag, params.SourceType, params.Version, params.Config, params.WorkflowID)
	return resp, err
}

// SyncActivity runs the sync command to transfer data between source and destination
func SyncActivity(ctx context.Context, params *SyncParams) (map[string]interface{}, error) {
	// Get activity logger
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity",
		"jobId", params.JobID,
		"workflowID", params.WorkflowID)

	// Get job details from database
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(params.JobID)
	if err != nil {
		logger.Error("Failed to get job details",
			"error", err,
			"job_id", params.JobID,
			"workflow_id", params.WorkflowID)
	} else if job == nil {
		logger.Error("Job not found in database",
			"job_id", params.JobID,
			"workflow_id", params.WorkflowID)
	}

	// Track sync start event
	properties := map[string]interface{}{
		"job_id":      params.JobID,
		"workflow_id": params.WorkflowID,
		"started_at":  time.Now().UTC().Format(time.RFC3339),
	}
	if job != nil {
		properties["job_name"] = job.Name
		// Get username from UserORM if CreatedBy exists
		if job.CreatedBy != nil {
			userORM := database.NewUserORM()
			if fullUser, err := userORM.GetByID(job.CreatedBy.ID); err == nil {
				properties["created_by"] = fullUser.Username
			}
		}
		properties["created_at"] = job.CreatedAt.Format(time.RFC3339)
		if job.SourceID != nil {
			properties["source_type"] = job.SourceID.Type
			properties["source_name"] = job.SourceID.Name
		}
		if job.DestID != nil {
			properties["destination_type"] = job.DestID.DestType
			properties["destination_name"] = job.DestID.Name
		}
	}

	if err := telemetry.TrackEvent(ctx, constants.EventSyncStarted, properties); err != nil {
		logger.Error("Failed to track sync start event", "error", err)
	}

	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())
	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Running sync command")
	// Execute the sync operation
	result, err := runner.RunSync(
		ctx,
		params.JobID,
		params.WorkflowID,
	)
	if err != nil {
		logger.Error("Sync command failed", "error", err)
		return result, fmt.Errorf("sync command failed: %v", err)
	}

	return result, nil
}

// TrackEventActivity tracks workflow events
func TrackEventActivity(ctx context.Context, eventName string, properties map[string]interface{}) error {
	return telemetry.TrackEvent(ctx, eventName, properties)
}
