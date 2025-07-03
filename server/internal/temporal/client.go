package temporal

import (
	"context"
	"fmt"
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

// SyncAction represents the type of action to perform
type SyncAction string

const (
	ActionCreate  SyncAction = "create"
	ActionUpdate  SyncAction = "update"
	ActionDelete  SyncAction = "delete"
	ActionTrigger SyncAction = "trigger"
	ActionPause   SyncAction = "pause"
	ActionUnpause SyncAction = "unpause"
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
func (c *Client) GetClient() client.Client {
	return c.temporalClient
}

// GetCatalog runs a workflow to discover catalog data
func (c *Client) GetCatalog(ctx context.Context, sourceType, version, config, streamsConfig string) (map[string]interface{}, error) {
	params := &ActivityParams{
		SourceType:    sourceType,
		Version:       version,
		Config:        config,
		WorkflowID:    fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().Unix()),
		Command:       docker.Discover,
		StreamsConfig: streamsConfig,
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

// ManageSync handles all sync operations (create, update, delete, trigger)
func (c *Client) ManageSync(ctx context.Context, projectID string, jobID int, frequency string, action SyncAction) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("sync-%s-%d", projectID, jobID)
	scheduleID := fmt.Sprintf("schedule-%s", workflowID)

	handle := c.temporalClient.ScheduleClient().GetHandle(ctx, scheduleID)
	currentSchedule, err := handle.Describe(ctx)
	scheduleExists := err == nil
	if action != ActionCreate && !scheduleExists {
		return nil, fmt.Errorf("schedule does not exist")
	}
	switch action {
	case ActionCreate:
		if frequency == "" {
			return nil, fmt.Errorf("frequency is required for creating schedule")
		}
		if scheduleExists {
			return nil, fmt.Errorf("schedule already exists")
		}
		return c.createSchedule(ctx, handle, scheduleID, workflowID, frequency, jobID)

	case ActionUpdate:
		if frequency == "" {
			return nil, fmt.Errorf("frequency is required for updating schedule")
		}
		return c.updateSchedule(ctx, handle, currentSchedule, scheduleID, frequency)

	case ActionDelete:
		if err := handle.Delete(ctx); err != nil {
			return nil, fmt.Errorf("failed to delete schedule: %s", err)
		}
		return map[string]interface{}{"message": "Schedule deleted successfully"}, nil

	case ActionTrigger:
		if err := handle.Trigger(ctx, client.ScheduleTriggerOptions{
			Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
		}); err != nil {
			return nil, fmt.Errorf("failed to trigger schedule: %s", err)
		}
		return map[string]interface{}{"message": "Schedule triggered successfully"}, nil
	case ActionPause:
		if err := handle.Pause(ctx, client.SchedulePauseOptions{
			Note: "Paused via API",
		}); err != nil {
			return nil, fmt.Errorf("failed to pause schedule: %s", err)
		}
		return map[string]interface{}{"message": "Schedule paused successfully"}, nil

	case ActionUnpause:
		if err := handle.Unpause(ctx, client.ScheduleUnpauseOptions{
			Note: "Unpaused via API",
		}); err != nil {
			return nil, fmt.Errorf("failed to unpause schedule: %s", err)
		}
		return map[string]interface{}{"message": "Schedule unpaused successfully"}, nil

	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

// createSchedule creates a new schedule
func (c *Client) createSchedule(ctx context.Context, _ client.ScheduleHandle, scheduleID, workflowID, frequency string, jobID int) (map[string]interface{}, error) {
	cronSpec := utils.ToCron(frequency)

	_, err := c.temporalClient.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{cronSpec},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			Workflow:  RunSyncWorkflow,
			Args:      []any{jobID},
			TaskQueue: TaskQueue,
		},
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %s", err)
	}

	return map[string]interface{}{
		"message": "Schedule created successfully",
		"cron":    cronSpec,
	}, nil
}

// updateSchedule updates an existing schedule
func (c *Client) updateSchedule(ctx context.Context, handle client.ScheduleHandle, currentSchedule *client.ScheduleDescription, _, frequency string) (map[string]interface{}, error) {
	cronSpec := utils.ToCron(frequency)

	// Check if update is needed
	if len(currentSchedule.Schedule.Spec.CronExpressions) > 0 &&
		currentSchedule.Schedule.Spec.CronExpressions[0] == cronSpec {
		return map[string]interface{}{"message": "Schedule already up to date"}, nil
	}

	err := handle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			input.Description.Schedule.Spec = &client.ScheduleSpec{
				CronExpressions: []string{cronSpec},
			}
			return &client.ScheduleUpdate{
				Schedule: &input.Description.Schedule,
			}, nil
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %s", err)
	}
	return map[string]interface{}{
		"message": "Schedule updated successfully",
		"cron":    cronSpec,
	}, nil
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
