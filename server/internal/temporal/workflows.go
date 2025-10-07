package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Retry policy constants
var (
	// DefaultRetryPolicy is used for standard operations like discovery and testing connections
	DefaultRetryPolicy = &temporal.RetryPolicy{
		InitialInterval:    time.Second * 15,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute * 10,
		MaximumAttempts:    1,
	}
)

// DiscoverCatalogWorkflow is a workflow for discovering catalogs
func DiscoverCatalogWorkflow(ctx workflow.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Execute the DiscoverCatalogActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, DiscoverCatalogActivity, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FetchSpecWorkflow is a workflow for fetching connector specifications
func FetchSpecWorkflow(ctx workflow.Context, params *ActivityParams) (models.SpecOutput, error) {
	// Execute the FetchSpecActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		HeartbeatTimeout:    time.Minute * 1,
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result models.SpecOutput
	err := workflow.ExecuteActivity(ctx, FetchSpecActivity, params).Get(ctx, &result)
	if err != nil {
		return models.SpecOutput{}, err
	}

	return result, nil
}

// TestConnectionWorkflow is a workflow for testing connections
func TestConnectionWorkflow(ctx workflow.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Execute the TestConnectionActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, TestConnectionActivity, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RunSyncWorkflow is a workflow for running data synchronization
func RunSyncWorkflow(ctx workflow.Context, jobID int) (result map[string]interface{}, err error) {
	logger := workflow.GetLogger(ctx)
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 24 * 30, // 30 days
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 15,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 10,
			MaximumAttempts:    0,
		},
		WaitForCancellation: true,
		HeartbeatTimeout:    time.Minute * 1,
	}
	params := SyncParams{
		JobID:      jobID,
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	// Defer cleanup for cancellation
	defer func() {
		logger.Info("executing workflow cleanup...")
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		cleanupOptions := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute * 15,
			RetryPolicy:         DefaultRetryPolicy,
		}
		newCtx = workflow.WithActivityOptions(newCtx, cleanupOptions)
		perr := workflow.ExecuteActivity(newCtx, SyncCleanupActivity, params).Get(newCtx, nil)
		if perr != nil {
			perr = fmt.Errorf("cleanup error: %s", perr)
			if err != nil {
				// preserve original err, just append cleanup info
				err = fmt.Errorf("%s; cleanup error: %s", err, perr)
			}
		}
	}()

	err = workflow.ExecuteActivity(ctx, SyncActivity, params).Get(ctx, &result)
	if err != nil {
		// Track sync failure event
		telemetry.TrackSyncFailed(context.Background(), jobID, params.WorkflowID)
		return nil, err
	}

	// Track sync completion
	telemetry.TrackSyncCompleted(context.Background(), jobID, params.WorkflowID)
	return result, nil
}
