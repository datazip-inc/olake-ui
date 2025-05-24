package temporal

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/datazip/olake-server/internal/docker"
	"go.temporal.io/api/enums/v1"
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
func (c *Client) GetCatalog(ctx context.Context, sourceType, version, config string) (map[string]interface{}, error) {
	params := ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		WorkflowID: fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().Unix()),
		Command:    docker.Discover,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
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
		WorkflowID: fmt.Sprintf("get-spec-%s-%d-%d", sourceType, sourceID, time.Now().Unix()),
		Command:    docker.Spec,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
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
		WorkflowID: fmt.Sprintf("test-connection-%s-%d-%d", sourceType, sourceID, time.Now().Unix()),
		Command:    docker.Check,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
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
func (c *Client) RunSync(ctx context.Context, sourceType, version, frequency, sourceConfig, destConfig, stateConfig, streamsConfig string,
	ProjectID, JobId, sourceID, destID int,
) (map[string]interface{}, error) {

	params := SyncParams{
		SourceType:    sourceType,
		Version:       version,
		SourceConfig:  sourceConfig,
		DestConfig:    destConfig,
		StateConfig:   stateConfig,
		StreamsConfig: streamsConfig,
		ProjectID:     ProjectID,
		JobId:         JobId,
		SourceID:      sourceID,
		DestID:        destID,
	}

	// Run the workflow immediately first
	exec, err := c.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("sync-%d-%d-%d-%d-%d", ProjectID, JobId, sourceID, destID, time.Now().Unix()),
		TaskQueue: TaskQueue,
	}, RunSyncWorkflow, params)

	if err != nil {
		return nil, fmt.Errorf("failed to start immediate sync workflow: %w", err)
	}

	// Only set up the schedule if frequency is specified
	if frequency != "" {
		cronSpec := toCron(frequency)
		id := fmt.Sprintf("sync-%d-%d-%d-%d", ProjectID, JobId, sourceID, destID)
		schedule := client.ScheduleSpec{
			CronExpressions: []string{cronSpec},
		}

		action := &client.ScheduleWorkflowAction{
			ID:        id,
			Workflow:  RunSyncWorkflow,
			Args:      []interface{}{params},
			TaskQueue: TaskQueue,
		}

		policies := client.SchedulePolicies{
			Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP, // Prevents overlapping workflow runs
		}
		scheduleID := fmt.Sprintf("schedule-%s", id)
		_, err = c.temporalClient.ScheduleClient().Create(ctx, client.ScheduleOptions{
			ID:      scheduleID,
			Spec:    schedule,
			Action:  action,
			Overlap: policies.Overlap,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create schedule: %w", err)
		}

		return map[string]interface{}{
			"message":         "Immediate sync started and schedule created with no overlap",
			"workflowRunId":   exec.GetRunID(),
			"scheduleCreated": true,
		}, nil
	}

	return map[string]interface{}{
		"message":         "Immediate sync started (no schedule)",
		"workflowRunId":   exec.GetRunID(),
		"scheduleCreated": false,
	}, nil
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

//		return &ListWorkflowExecutionsResponse{
//			Executions: executions,
//		}, nil
//	}
func toCron(frequency string) string {
	parts := strings.Split(strings.ToLower(frequency), "-")
	if len(parts) != 2 {
		return ""
	}

	valueStr, unit := parts[0], parts[1]
	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return ""
	}

	switch unit {
	case "minutes":
		return fmt.Sprintf("*/%d * * * *", value) // Every N minutes
	case "hours":
		return fmt.Sprintf("0 */%d * * *", value) // Every N hours at minute 0
	case "days":
		return fmt.Sprintf("0 0 */%d * *", value) // Every N days at midnight
	case "weeks":
		// Every N weeks on Sunday (0), cron doesn't support "every N weeks" directly,
		// so simulate with day-of-week field (best-effort)
		return fmt.Sprintf("0 0 * * */%d", value)
	case "months":
		return fmt.Sprintf("0 0 1 */%d *", value) // Every N months on the 1st at midnight
	default:
		return ""
	}
}
