package workflows

import (
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"olake-k8s-worker/activities"
	"olake-k8s-worker/shared"
)

// No timeout retry policy for long-running operations
var DefaultRetryPolicy = &temporal.RetryPolicy{
	MaximumAttempts: 1, // No retries for data operations
}

// DiscoverCatalogWorkflow is a workflow for discovering catalogs using K8s Jobs
func DiscoverCatalogWorkflow(ctx workflow.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		// No StartToCloseTimeout - let it run as long as needed
		RetryPolicy: DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, activities.DiscoverCatalogActivity, params).Get(ctx, &result)
	return result, err
}

// TestConnectionWorkflow is a workflow for testing connections using K8s Jobs
func TestConnectionWorkflow(ctx workflow.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		// No timeout - connection tests can also be slow for large databases
		RetryPolicy: DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, activities.TestConnectionActivity, params).Get(ctx, &result)
	return result, err
}

// RunSyncWorkflow is a workflow for running data synchronization using K8s Jobs
func RunSyncWorkflow(ctx workflow.Context, jobID int) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		// No timeout - sync operations can take hours for large datasets
		RetryPolicy: DefaultRetryPolicy,
	}
	params := shared.SyncParams{
		JobID:      jobID,
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, activities.SyncActivity, params).Get(ctx, &result)
	return result, err
}
