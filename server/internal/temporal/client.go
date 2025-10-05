package temporal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"golang.org/x/mod/semver"
)

var (
	TemporalAddress string
)

// SyncAction represents the type of action to perform
type SyncAction string

// Command represents the command to execute
type Command string

// ExecutionRequest is the request body for the workflow
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

const (
	// TaskQueue is the default task queue for Olake Docker workflows
	TaskQueue = "OLAKE_DOCKER_TASK_QUEUE"

	ActionCreate  SyncAction = "create"
	ActionUpdate  SyncAction = "update"
	ActionDelete  SyncAction = "delete"
	ActionTrigger SyncAction = "trigger"
	ActionPause   SyncAction = "pause"
	ActionUnpause SyncAction = "unpause"

	Discover Command = "discover"
	Check    Command = "check"
	Sync     Command = "sync"
	Spec     Command = "spec"
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

func (c *Client) Close() {
	if c.temporalClient != nil {
		c.temporalClient.Close()
	}
}

func (c *Client) GetClient() client.Client {
	return c.temporalClient
}

func (c *Client) GetCatalog(ctx context.Context, jobName, sourceType, version, config, streamsConfig string) (map[string]interface{}, error) {
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

	if jobName != "" && semver.Compare(version, "v0.2.0") >= 0 {
		cmdArgs = append(cmdArgs, "--destination-database-prefix", jobName)
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
		Timeout:       GetWorkflowTimeout(Discover),
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

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	return result, nil
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
		Timeout:       GetWorkflowTimeout(Check),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "ExecuteWorkflow", req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test connection workflow: %v", err)
	}

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	connectionStatus, ok := result["connectionStatus"].(map[string]interface{})
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

// FetchSpec runs a workflow to fetch connector specifications
func (c *Client) FetchSpec(ctx context.Context, destinationType, sourceType, version string) (dto.SpecOutput, error) {
	workflowID := fmt.Sprintf("fetch-spec-%s-%d", sourceType, time.Now().Unix())

	// spec version >= DefaultSpecVersion is required
	if semver.Compare(version, constants.DefaultSpecVersion) < 0 {
		version = constants.DefaultSpecVersion
	}

	cmdArgs := []string{
		"spec",
	}
	if destinationType != "" {
		cmdArgs = append(cmdArgs, "--destination-type", destinationType)
	}

	req := &ExecutionRequest{
		Type:          "docker",
		Command:       Spec,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       nil,
		WorkflowID:    workflowID,
		JobID:         0,
		Timeout:       GetWorkflowTimeout(Spec),
		OutputFile:    "",
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "ExecuteWorkflow", req)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to execute discover workflow: %v", err)
	}

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	return dto.SpecOutput{
		Spec: result,
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
			Workflow:  "ExecuteSyncWorkflow",
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
