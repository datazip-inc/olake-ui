package temporal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/utils"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// TaskQueue is the default task queue for Olake Docker workflows
const TaskQueue = "OLAKE_DOCKER_TASK_QUEUE"

var (
	TemporalAddress string
)

func init() {
	TemporalAddress = web.AppConfig.DefaultString("TEMPORAL_ADDRESS", "localhost:7233")
}

// Client provides methods to interact with Temporal
type Client struct {
	temporalClient client.Client
}

// NewClient creates a new Temporal client
func NewClient() (*Client, error) {
	c, err := client.Dial(client.Options{
		HostPort: TemporalAddress,
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
	params := &ActivityParams{
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

// TestConnection runs a workflow to test connection
func (c *Client) TestConnection(ctx context.Context, flag, sourceType, version, config string) (map[string]interface{}, error) {
	params := &ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		WorkflowID: fmt.Sprintf("test-connection-%s-%d", sourceType, time.Now().Unix()),
		Command:    docker.Check,
		Flag:       flag,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, TestConnectionWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test connection workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	return result, nil
}

// CreateSync creates a sync workflow
func (c *Client) CreateSync(ctx context.Context, frequency, projectIDStr string, jobID int, runImmediately bool) (map[string]interface{}, error) {
	id := fmt.Sprintf("sync-%s-%d", projectIDStr, jobID)
	scheduleID := fmt.Sprintf("schedule-%s", id)
	scheduleHandle := c.temporalClient.ScheduleClient().GetHandle(ctx, scheduleID)
	currentSchedule, err := scheduleHandle.Describe(ctx)
	scheduleExists := err == nil

	// Handle schedule creation/update
	if frequency != "" && !runImmediately {
		if err := c.handleSchedule(ctx, scheduleHandle, frequency, scheduleExists, currentSchedule, id, scheduleID, jobID); err != nil {
			return nil, err
		}
		return map[string]interface{}{"message": "Schedule created/updated successfully"}, nil
	}

	// Handle immediate run
	if runImmediately {
		if !scheduleExists && frequency == "" {
			return nil, fmt.Errorf("cannot run immediately without a schedule - frequency must be specified or existing schedule must be present")
		}

		err = scheduleHandle.Trigger(ctx, client.ScheduleTriggerOptions{
			Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to trigger schedule manually: %s", err)
		}

		return map[string]interface{}{
			"message": "sync triggered successfully",
		}, nil
	}

	return nil, fmt.Errorf("no action taken: either frequency must be provided for scheduling or runImmediately must be true")
}

// handleSchedule creates or updates a schedule based on frequency
func (c *Client) handleSchedule(ctx context.Context, scheduleHandle client.ScheduleHandle, frequency string, scheduleExists bool, currentSchedule *client.ScheduleDescription, id, scheduleID string, jobID int) error {
	cronSpec := utils.ToCron(frequency)

	if scheduleExists {
		// Check if frequency has changed
		needsUpdate := len(currentSchedule.Schedule.Spec.CronExpressions) == 0 ||
			currentSchedule.Schedule.Spec.CronExpressions[0] != cronSpec

		if needsUpdate {
			// Update existing schedule
			log.Printf("Updating schedule %s to cron: %s", scheduleID, cronSpec)
			updateSchedule := func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
				input.Description.Schedule.Spec = &client.ScheduleSpec{
					CronExpressions: []string{cronSpec},
				}
				return &client.ScheduleUpdate{
					Schedule: &input.Description.Schedule,
				}, nil
			}
			err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
				DoUpdate: updateSchedule,
			})
			if err != nil {
				return fmt.Errorf("failed to update schedule: %s", err)
			}
		}
	} else {
		// Create new schedule
		schedule := client.ScheduleSpec{
			CronExpressions: []string{cronSpec},
		}
		action := &client.ScheduleWorkflowAction{
			ID:        id,
			Workflow:  RunSyncWorkflow,
			Args:      []any{jobID},
			TaskQueue: TaskQueue,
		}

		_, err := c.temporalClient.ScheduleClient().Create(ctx, client.ScheduleOptions{
			ID:      scheduleID,
			Spec:    schedule,
			Action:  action,
			Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
		})
		if err != nil {
			return fmt.Errorf("failed to create schedule: %s", err)
		}
	}

	return nil
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
