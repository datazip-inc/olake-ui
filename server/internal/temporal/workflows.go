package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Retry policy constants
var (
	// DefaultRetryPolicy is used for standard operations like discovery and testing connections
	DefaultRetryPolicy = &temporal.RetryPolicy{
		InitialInterval:    time.Second * 5,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute * 5,
		MaximumAttempts:    1,
	}
)

// DockerRunnerWorkflow orchestrates the Docker command execution as a workflow
// func DockerRunnerWorkflow(ctx workflow.Context, params ActivityParams) (map[string]interface{}, error) {
// 	options := workflow.ActivityOptions{
// 		StartToCloseTimeout: time.Minute * 5,
// 		RetryPolicy: &temporal.RetryPolicy{ // <- Use temporal.RetryPolicy
// 			InitialInterval:    time.Second,
// 			BackoffCoefficient: 2.0,
// 			MaximumInterval:    time.Minute,
// 			MaximumAttempts:    1,
// 		},
// 	}
// 	ctx = workflow.WithActivityOptions(ctx, options)

// 	var result map[string]interface{}
// 	err := workflow.ExecuteActivity(ctx, ExecuteDockerCommandActivity, params).Get(ctx, &result)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// DiscoverCatalogWorkflow is a workflow for discovering catalogs
func DiscoverCatalogWorkflow(ctx workflow.Context, params ActivityParams) (map[string]interface{}, error) {
	// Execute the DiscoverCatalogActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
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

// GetSpecWorkflow is a workflow for getting connector specs
// func GetSpecWorkflow(ctx workflow.Context, params ActivityParams) (map[string]interface{}, error) {
// 	// Execute the GetSpecActivity directly
// 	options := workflow.ActivityOptions{
// 		StartToCloseTimeout: time.Minute * 5,
// 		RetryPolicy: DefaultRetryPolicy,
// 	}
// 	ctx = workflow.WithActivityOptions(ctx, options)

// 	var result map[string]interface{}
// 	err := workflow.ExecuteActivity(ctx, GetSpecActivity, params).Get(ctx, &result)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// TestConnectionWorkflow is a workflow for testing connections
func TestConnectionWorkflow(ctx workflow.Context, params ActivityParams) (map[string]interface{}, error) {
	// Execute the TestConnectionActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
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
		StartToCloseTimeout: time.Minute * 15, // Longer timeout for sync operations
		RetryPolicy:         DefaultRetryPolicy,
	}
	params := SyncParams{
		JobId:      jobID,
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, SyncActivity, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}