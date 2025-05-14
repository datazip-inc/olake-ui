package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip/olake-server/internal/docker"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// TaskQueue is the default task queue for Olake Docker workflows
const TaskQueue = "OLAKE_DOCKER_TASK_QUEUE"

// Client provides methods to interact with Temporal
type Client struct {
	temporalClient client.Client
}

// NewClient creates a new Temporal client
func NewClient(address string) (*Client, error) {
	if address == "" {
		address = "localhost:7233" // Default Temporal address
	}

	c, err := client.Dial(client.Options{
		HostPort: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %v", err)
	}

	return &Client{
		temporalClient: c,
	}, nil
}

// Close closes the Temporal client
func (c *Client) Close() {
	if c.temporalClient != nil {
		c.temporalClient.Close()
	}
}

// GetCatalog runs a workflow to discover catalog data
func (c *Client) GetCatalog(ctx context.Context, sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	params := ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		SourceID:   sourceID,
		Command:    docker.Discover,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("discover-catalog-%s-%d-%d", sourceType, sourceID, time.Now().Unix()),
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, DiscoverCatalogWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute discover workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	return result, nil
}

// GetSpec runs a workflow to get connector specification
func (c *Client) GetSpec(ctx context.Context, sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	params := ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		SourceID:   sourceID,
		Command:    docker.Spec,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("get-spec-%s-%d-%d", sourceType, sourceID, time.Now().Unix()),
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, GetSpecWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute spec workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	return result, nil
}

// TestConnection runs a workflow to test connection
func (c *Client) TestConnection(ctx context.Context, sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	params := ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		SourceID:   sourceID,
		Command:    docker.Check,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("test-connection-%s-%d-%d", sourceType, sourceID, time.Now().Unix()),
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, TestConnectionWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute connection test workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	return result, nil
}

// RunSync runs a workflow to sync data between source and destination
func (c *Client) RunSync(ctx context.Context, sourceType, version, sourceConfig, destConfig, streamsConfig string, ProjectID, JobId, sourceID, destID int) (map[string]interface{}, error) {
	params := SyncParams{
		SourceType:    sourceType,
		Version:       version,
		SourceConfig:  sourceConfig,
		DestConfig:    destConfig,
		StreamsConfig: streamsConfig,
		ProjectID:     ProjectID,
		JobId:         JobId,
		SourceID:      sourceID,
		DestID:        destID,
		WorkflowID:    fmt.Sprintf("sync-%d-%d-%d-%d-%d", ProjectID, JobId, sourceID, destID, time.Now().Unix()),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: TaskQueue,
	}
	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, RunSyncWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sync workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	return result, nil
}

// WorkflowExecution represents information about a workflow execution
type WorkflowExecution struct {
	ID            string
	RunID         string
	Type          string
	StartTime     time.Time
	ExecutionTime time.Time
	CloseTime     time.Time
	Status        string
	HistoryLength int64
	LogFolderName string
}

// ListWorkflowExecutionsRequest represents a request to list workflow executions
type ListWorkflowExecutionsRequest struct {
	Query string
}

// ListWorkflowExecutionsResponse represents the response from listing workflow executions
type ListWorkflowExecutionsResponse struct {
	Executions []WorkflowExecution
}

// ListWorkflow lists workflow executions based on the provided query
func (c *Client) ListWorkflow(ctx context.Context, request *workflowservice.ListWorkflowExecutionsRequest) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	// Query workflows using the SDK's ListWorkflow method
	resp, err := c.temporalClient.ListWorkflow(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing workflow executions: %v", err)
	}

	return resp, nil
}

// GetWorkflow retrieves a workflow by ID and runID
func (c *Client) GetWorkflow(ctx context.Context, workflowID, runID string) client.WorkflowRun {
	return c.temporalClient.GetWorkflow(ctx, workflowID, runID)
}

// 		execution := WorkflowExecution{
// 			ID:    exec.Execution.WorkflowId,
// 			RunID: exec.Execution.RunId,
// 			Type:  exec.Type.Name,
// 		}

// 		// Convert timestamps if available
// 		if exec.StartTime != nil {
// 			execution.StartTime = exec.StartTime.AsTime()
// 		}
// 		if exec.ExecutionTime != nil {
// 			execution.ExecutionTime = exec.ExecutionTime.AsTime()
// 		}
// 		if exec.CloseTime != nil {
// 			execution.CloseTime = exec.CloseTime.AsTime()
// 		}

// 		// Add status and history length
// 		execution.Status = exec.Status.String()
// 		execution.HistoryLength = exec.HistoryLength

// 		executions = append(executions, execution)
// 	}

// 	return &ListWorkflowExecutionsResponse{
// 		Executions: executions,
// 	}, nil
// }
