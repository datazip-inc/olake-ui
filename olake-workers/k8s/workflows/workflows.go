package workflows

import (
	"strconv"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"github.com/spf13/viper"

	"olake-ui/olake-workers/k8s/shared"
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

// parseTimeout parses a timeout from viper with fallback
func parseTimeout(envKey string, defaultValue time.Duration) time.Duration {
	timeoutStr := viper.GetString(envKey)
	if timeoutStr == "" {
		return defaultValue
	}

	if seconds, err := strconv.Atoi(timeoutStr); err == nil {
		return time.Duration(seconds) * time.Second
	}

	if duration, err := time.ParseDuration(timeoutStr); err == nil {
		return duration
	}

	return defaultValue
}

// getActivityTimeout reads activity timeout from viper configuration
func getActivityTimeout(operation string) time.Duration {
	switch operation {
	case "discover":
		return parseTimeout("timeouts.activity.discover", 2*time.Hour)
	case "test":
		return parseTimeout("timeouts.activity.test", 2*time.Hour)
	case "sync":
		return parseTimeout("timeouts.activity.sync", 700*time.Hour)
	default:
		return 30 * time.Minute
	}
}
