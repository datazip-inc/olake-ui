package temporal

import (
	"context"
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
func RunSyncWorkflow(ctx workflow.Context, jobID int) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		// Using large duration (e.g., 10 years)
		StartToCloseTimeout: time.Hour * 24 * 30, // 30 days
		RetryPolicy:         DefaultRetryPolicy,
	}
	params := SyncParams{
		JobID:      jobID,
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, SyncActivity, params).Get(ctx, &result)
	if err != nil {
		// Track sync failure event
		telemetry.TrackSyncFailed(context.Background(), jobID, params.WorkflowID)
		return nil, err
	}

	// Track sync completion
	telemetry.TrackSyncCompleted(context.Background(), jobID, params.WorkflowID)
	return result, nil
}
