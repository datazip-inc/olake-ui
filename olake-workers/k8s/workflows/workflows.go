package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"olake-ui/olake-workers/k8s/shared"
	"olake-ui/olake-workers/k8s/utils/helpers"
)

// Retry policy matching server-side configuration
var DefaultRetryPolicy = &temporal.RetryPolicy{
	InitialInterval:    time.Second * 5,
	BackoffCoefficient: 2.0,
	MaximumInterval:    time.Minute * 5,
	MaximumAttempts:    1,
}

// DiscoverCatalogWorkflow is a workflow for discovering catalogs using K8s Jobs
func DiscoverCatalogWorkflow(ctx workflow.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: helpers.GetActivityTimeout("discover"),
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, "DiscoverCatalogActivity", params).Get(ctx, &result)
	return result, err
}

// TestConnectionWorkflow is a workflow for testing connections using K8s Jobs
func TestConnectionWorkflow(ctx workflow.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: helpers.GetActivityTimeout("test"),
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, "TestConnectionActivity", params).Get(ctx, &result)
	return result, err
}

// RunSyncWorkflow is a workflow for running data synchronization using K8s Jobs
func RunSyncWorkflow(ctx workflow.Context, jobID int) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: helpers.GetActivityTimeout("sync"),
		RetryPolicy:         DefaultRetryPolicy,
	}
	params := shared.SyncParams{
		JobID:      jobID,
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, "SyncActivity", params).Get(ctx, &result)
	return result, err
}

