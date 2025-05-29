package temporal

import (
	"time"

	"github.com/datazip/olake-server/internal/docker"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DockerCommandParams contains parameters for Docker commands (legacy)
type DockerCommandParams struct {
	SourceType string
	Version    string
	Config     string
	SourceID   int
	Command    string
}

// ActivityParams contains parameters for Docker command activities
type ActivityParams struct {
	SourceType   string
	Version      string
	Config       string
	SourceID     int
	Command      docker.Command
	DestConfig   string
	DestID       int
	WorkflowID   string
	StreamConfig string
	Flag         string
}

// SyncParams contains parameters for sync activities
type SyncParams struct {
	SourceType    string
	Version       string
	SourceConfig  string
	DestConfig    string
	StreamsConfig string
	StateConfig   string
	JobId         int
	ProjectID     int
	SourceID      int
	DestID        int
	WorkflowID    string
}

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
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    1,
		},
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
// 		RetryPolicy: &temporal.RetryPolicy{
// 			InitialInterval:    time.Second,
// 			BackoffCoefficient: 2.0,
// 			MaximumInterval:    time.Minute,
// 			MaximumAttempts:    1,
// 		},
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
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    1,
		},
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
func RunSyncWorkflow(ctx workflow.Context, params SyncParams) (map[string]interface{}, error) {
	// Get workflow info from context
	workflowInfo := workflow.GetInfo(ctx)
	workflowID := workflowInfo.WorkflowExecution.ID
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 15, // Longer timeout for sync operations
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 5,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    1, // Fewer retries for sync as it's more expensive
		},
	}
	params.WorkflowID = workflowID
	ctx = workflow.WithActivityOptions(ctx, options)
	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, SyncActivity, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
