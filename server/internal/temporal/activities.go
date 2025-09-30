package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/internal/models"
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
		params.JobName,
	)
	if err != nil {
		logger.Error("Sync command failed", "error", err)
		return result, fmt.Errorf("sync command failed: %v", err)
	}

	return result, nil
}

// FetchSpecActivity runs the spec command to get connector specifications
func FetchSpecActivity(ctx context.Context, params *ActivityParams) (models.SpecOutput, error) {
	runner := docker.NewRunner(docker.GetDefaultConfigDir())
	return runner.FetchSpec(ctx, params.DestinationType, params.SourceType, params.Version, params.WorkflowID)
}

// TestConnectionActivity runs the check command to test connection
func TestConnectionActivity(ctx context.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())
	resp, err := runner.TestConnection(ctx, params.Flag, params.SourceType, params.Version, params.Config, params.WorkflowID)
	return resp, err
}

// SyncActivity runs the sync command to transfer data between source and destination
func SyncActivity(ctx context.Context, params *SyncParams) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity", "jobId", params.JobID, "workflowID", params.WorkflowID)

	activity.RecordHeartbeat(ctx, "Running sync command")

	// Run the sync in a goroutine so we can react to cancel concurrently
	type resErr struct {
		res map[string]interface{}
		err error
	}
	done := make(chan resErr, 1)

	go func() {
		runner := docker.NewRunner(docker.GetDefaultConfigDir())
		res, err := runner.RunSync(ctx, params.JobID, params.WorkflowID)
		done <- resErr{res: res, err: err}
	}()

	for {
		select {
		case <-ctx.Done():
			// Cancellation observed: stop container by hashed workflowID, then return canceled
			_ = docker.StopContainer(context.Background(), params.WorkflowID) // best-effort; non-blocking on ctx cancel
			return nil, ctx.Err()
		case r := <-done:
			if r.err != nil {
				logger.Error("Sync command failed", "error", r.err)
				return r.res, fmt.Errorf("sync command failed: %v", r.err)
			}
			return r.res, nil
		default:
			// Keep heartbeating so cancel is delivered promptly
			activity.RecordHeartbeat(ctx, "sync in progress")
			time.Sleep(1 * time.Second)
		}
	}
}
