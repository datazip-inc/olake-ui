package temporal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
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

type Command string

var (
	Discover Command = "discover"
	Check    Command = "check"
	Sync     Command = "sync"
	Spec     Command = "spec"
)

// New worker types for the common worker
type ExecutionRequest struct {
	Type          string        `json:"type"`
	Command       Command       `json:"command"`
	ConnectorType string        `json:"connector_type"`
	Version       string        `json:"version"`
	Args          []string      `json:"args"`
	Configs       []JobConfig   `json:"configs"`
	WorkflowID    string        `json:"workflow_id"`
	JobID         int           `json:"job_id"`
	Timeout       time.Duration `json:"timeout"`
	OutputFile    string        `json:"output_file"` // to get the output file from the workflow
}

type JobConfig struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

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

func (c *Client) Close() {
	if c.temporalClient != nil {
		c.temporalClient.Close()
	}
}

func GetTimeout(op Command) time.Duration {
	switch op {
	case Discover:
		return time.Minute * 10
	case Check:
		return time.Minute * 10
	case Spec:
		return time.Minute * 5
	case Sync:
		return time.Hour * 24 * 30
	// check what can the fallback time be
	default:
		return time.Minute * 5
	}

}

func (c *Client) GetClient() client.Client {
	return c.temporalClient
}

func (c *Client) GetCatalog(ctx context.Context, sourceType, version, config, streamsConfig string) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().Unix())

	configs := []JobConfig{
		{Name: "config.json", Data: config},
		{Name: "streams.json", Data: streamsConfig},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID()},
	}

	cmdArgs := []string{
		"discover",
		"--config",
		"/mnt/config/config.json",
	}
	if streamsConfig != "" {
		cmdArgs = append(cmdArgs, "--catalog", "/mnt/config/streams.json")
	}
	if encryptionKey := os.Getenv(constants.EncryptionKey); encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Type:          "docker",
		Command:       Discover,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       configs,
		WorkflowID:    workflowID,
		JobID:         0,
		Timeout:       GetTimeout(Discover),
		OutputFile:    "streams.json",
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "ExecuteWorkflow", req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute discover workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from worker")
	}

	logResult, err := utils.ExtractJSON(response)
	if err != nil {
		return nil, err
	}

	return logResult, nil
}

func (c *Client) TestConnection(ctx context.Context, flag, sourceType, version, config string) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("test-connection-%s-%d", sourceType, time.Now().Unix())

	configs := []JobConfig{
		{Name: "config.json", Data: config},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID()},
	}

	cmdArgs := []string{
		"check",
		fmt.Sprintf("--%s", flag),
		"/mnt/config/config.json",
	}
	if encryptionKey := os.Getenv(constants.EncryptionKey); encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Type:          "docker",
		Command:       Check,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       configs,
		WorkflowID:    workflowID,
		Timeout:       GetTimeout(Check),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "ExecuteWorkflow", req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test connection workflow: %v", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from worker")
	}

	logMsg, err := utils.ExtractJSON(response)
	if err != nil {
		return nil, err
	}

	connectionStatus, ok := logMsg["connectionStatus"].(map[string]interface{})
	if !ok || connectionStatus == nil {
		return nil, fmt.Errorf("connection status not found")
	}

	status, statusOk := connectionStatus["status"].(string)
	message, _ := connectionStatus["message"].(string) // message is optional
	if !statusOk {
		return nil, fmt.Errorf("connection status not found")
	}

	return map[string]interface{}{
		"message": message,
		"status":  status,
	}, nil
}

func (c *Client) ManageSync(ctx context.Context, job *models.Job, action SyncAction) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("sync-%s-%d", job.ProjectID, job.ID)
	scheduleID := fmt.Sprintf("schedule-%s", workflowID)

	handle := c.temporalClient.ScheduleClient().GetHandle(ctx, scheduleID)
	_, err := handle.Describe(ctx)
	scheduleExists := err == nil
	if action != ActionCreate && !scheduleExists {
		return nil, fmt.Errorf("schedule does not exist")
	}
	switch action {
	case ActionCreate:
		if job.Frequency == "" {
			return nil, fmt.Errorf("frequency is required for creating schedule")
		}
		if scheduleExists {
			return nil, fmt.Errorf("schedule already exists")
		}
		return c.createSchedule(ctx, job, scheduleID, workflowID)

	case ActionUpdate:
		if job.Frequency == "" {
			return nil, fmt.Errorf("frequency is required for updating schedule")
		}
		return c.updateSchedule(ctx, handle, job)

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

func (c *Client) createSchedule(ctx context.Context, job *models.Job, scheduleID, workflowID string) (map[string]interface{}, error) {
	cronSpec := utils.ToCron(job.Frequency)

	req := buildExecutionReqForSync(job, workflowID)

	_, err := c.temporalClient.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{cronSpec},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			Workflow:  "ExecuteWorkflow",
			Args:      []any{req},
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

func (c *Client) updateSchedule(ctx context.Context, handle client.ScheduleHandle, job *models.Job) (map[string]interface{}, error) {
	cronSpec := utils.ToCron(job.Frequency)

	err := handle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			input.Description.Schedule.Spec = &client.ScheduleSpec{
				CronExpressions: []string{cronSpec},
			}

			workflowID := input.Description.Schedule.Action.(*client.ScheduleWorkflowAction).ID
			req := buildExecutionReqForSync(job, workflowID)

			input.Description.Schedule.Action = &client.ScheduleWorkflowAction{
				ID:        workflowID,
				Workflow:  "ExecuteSyncWorkflow",
				Args:      []any{req},
				TaskQueue: TaskQueue,
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

// buildExecutionReqForSync builds the ExecutionRequest for a sync job
func buildExecutionReqForSync(job *models.Job, workflowID string) ExecutionRequest {
	configs := []JobConfig{
		{Name: "source.json", Data: job.SourceID.Config},
		{Name: "destination.json", Data: job.DestID.Config},
		{Name: "streams.json", Data: job.StreamsConfig},
		{Name: "state.json", Data: job.State},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID()},
	}

	args := []string{
		"sync",
		"--config", "/mnt/config/source.json",
		"--destination", "/mnt/config/destination.json",
		"--catalog", "/mnt/config/streams.json",
		"--state", "/mnt/config/state.json",
	}

	return ExecutionRequest{
		Type:          "docker",
		Command:       Sync,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
		Args:          args,
		Configs:       configs,
		WorkflowID:    workflowID,
		JobID:         job.ID,
		Timeout:       GetTimeout(Sync),
		OutputFile:    "state.json",
	}
}
