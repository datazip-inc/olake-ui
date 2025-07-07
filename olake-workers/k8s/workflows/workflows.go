package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/config/helpers"
	"olake-ui/olake-workers/k8s/shared"
)

// Retry policy matching server-side configuration
var DefaultRetryPolicy = &temporal.RetryPolicy{
	InitialInterval:    time.Second * 5,
	BackoffCoefficient: 2.0,
	MaximumInterval:    time.Minute * 5,
	MaximumAttempts:    1,
}

// Global config instance (set during worker initialization)
var globalConfig *config.Config

// SetConfig sets the global configuration for workflows
func SetConfig(cfg *config.Config) {
	globalConfig = cfg
}

// getActivityTimeout returns activity timeout from config or fallback
func getActivityTimeout(operation string) time.Duration {
	if globalConfig != nil {
		return helpers.GetActivityTimeout(globalConfig, operation)
	}
	// Fallback defaults if config not available
	switch operation {
	case "discover":
		return time.Minute * 30
	case "test":
		return time.Minute * 30
	case "sync":
		return time.Hour * 4
	default:
		return time.Minute * 30
	}
}

// DiscoverCatalogWorkflow is a workflow for discovering catalogs using K8s Jobs
func DiscoverCatalogWorkflow(ctx workflow.Context, params *shared.ActivityParams) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: getActivityTimeout("discover"),
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
		StartToCloseTimeout: getActivityTimeout("test"),
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
		StartToCloseTimeout: getActivityTimeout("sync"),
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
