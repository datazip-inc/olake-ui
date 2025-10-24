package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/docker"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// DiscoverCatalogActivity runs the discover command to get catalog data
func DiscoverCatalogActivity(ctx context.Context, params *ActivityParams) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity",
		"sourceType", params.SourceType,
		"workflowID", params.WorkflowID)

	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.DefaultConfigDir)

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
		params.JobName,
	)
	if err != nil {
		logger.Error("Sync command failed", "error", err)
		return result, fmt.Errorf("sync command failed: %v", err)
	}

	return result, nil
}

// FetchSpecActivity runs the spec command to get conndriverector specifications
func FetchSpecActivity(ctx context.Context, params *ActivityParams) (dto.SpecOutput, error) {
	runner := docker.NewRunner(docker.GetDefaultConfigDir())
	return runner.FetchSpec(ctx, params.DestinationType, params.SourceType, params.Version, params.WorkflowID)
}

// TestConnectionActivity runs the check command to test connection
func TestConnectionActivity(ctx context.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.DefaultConfigDir)
	resp, err := runner.TestConnection(ctx, params.Flag, params.SourceType, params.Version, params.Config, params.WorkflowID)
	return resp, err
}

// SyncActivity runs the sync command to transfer data between source and destination
func SyncActivity(ctx context.Context, params *SyncParams) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity", "jobId", params.JobID, "workflowID", params.WorkflowID)

	activity.RecordHeartbeat(ctx, "Running sync command")

	type resErr struct {
		res map[string]interface{}
		err error
	}
	done := make(chan resErr, 1)
	// excueting sync in a goroutine to prevent blocking and monitoring the sync progress
	go func() {
		runner := docker.NewRunner(docker.GetDefaultConfigDir())
		res, err := runner.RunSync(ctx, params.JobID, params.WorkflowID)
		done <- resErr{res: res, err: err}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Info("SyncActivity canceled, deferring cleanup to SyncCleanupActivity")
			return nil, ctx.Err()
		case r := <-done:
			if r.err != nil {
				// CRITICAL: Check if error is because context was cancelled
				if ctx.Err() != nil {
					logger.Info("Goroutine failed due to context cancellation", "dockerError", r.err)
					return nil, ctx.Err() // Return cancellation error, not docker error
				}

				logger.Error("Sync command failed", "error", r.err)
				return r.res, temporal.NewNonRetryableApplicationError(r.err.Error(), "SyncFailed", r.err)
			}
			return r.res, nil
		default:
			activity.RecordHeartbeat(ctx, "sync in progress")
			time.Sleep(1 * time.Second)
		}
	}
}

// SyncCleanupActivity ensures container is fully stopped and state is persisted to database
func SyncCleanupActivity(ctx context.Context, params *SyncParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting cleanup activity", "jobId", params.JobID, "workflowID", params.WorkflowID)
	// Stop container gracefully
	logger.Info("Stopping container for cleanup %s", params.WorkflowID)
	if err := docker.StopContainer(ctx, params.WorkflowID); err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), "CleanupFailed", err)
	}
	runner := docker.NewRunner(docker.GetDefaultConfigDir())
	logger.Info("Persisting job state for workflowID %s", params.WorkflowID)
	if err := runner.PersistJobStateFromFile(params.JobID, params.WorkflowID); err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), "CleanupFailed", err)
	}
	logger.Info("Cleanup completed successfully")
	return nil
}
