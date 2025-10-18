package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/utils"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"golang.org/x/mod/semver"
)

// TaskQueue is the default task queue for Olake Docker workflows
const (
	DockerTaskQueue = "OLAKE_DOCKER_TASK_QUEUE"
	K8sTaskQueue    = "OLAKE_K8S_TASK_QUEUE"
)

var TaskQueue string

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

	// Choose task queue based on deployment mode
	deploymentMode := web.AppConfig.DefaultString("DEPLOYMENT_MODE", "docker")
	if deploymentMode == "kubernetes" {
		TaskQueue = K8sTaskQueue
	} else {
		TaskQueue = DockerTaskQueue
	}
}

// Client provides methods to interact with Temporal
type Temporal struct {
	Client client.Client
}

// NewClient creates a new Temporal client
func NewClient() (*Temporal, error) {
	c, err := client.Dial(client.Options{
		HostPort: TemporalAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %s", err)
	}

	return &Temporal{
		Client: c,
	}, nil
}

// Close closes the Temporal client
func (t *Temporal) Close() {
	if t.Client != nil {
		t.Client.Close()
	}
}

// GetCatalog runs a workflow to discover catalog data
func (t *Temporal) GetCatalog(ctx context.Context, sourceType, version, config, streamsConfig, jobName string) (map[string]interface{}, error) {
	params := &ActivityParams{
		SourceType:    sourceType,
		Version:       version,
		Config:        config,
		WorkflowID:    fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().Unix()),
		Command:       docker.Discover,
		StreamsConfig: streamsConfig,
		JobName:       jobName,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: TaskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, DiscoverCatalogWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute discover workflow: %s", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %s", err)
	}

	return result, nil
}

// FetchSpec runs a workflow to fetch connector specifications
func (t *Temporal) FetchSpec(ctx context.Context, destinationType, sourceType, version string) (dto.SpecOutput, error) {
	// spec version >= DefaultSpecVersion is required
	if semver.Compare(version, constants.DefaultSpecVersion) < 0 {
		version = constants.DefaultSpecVersion
	}

	params := &ActivityParams{
		SourceType:      sourceType,
		Version:         version,
		WorkflowID:      fmt.Sprintf("fetch-spec-%s-%d", sourceType, time.Now().Unix()),
		DestinationType: destinationType,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: TaskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, FetchSpecWorkflow, params)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to execute fetch spec workflow: %s", err)
	}

	var result dto.SpecOutput
	if err := run.Get(ctx, &result); err != nil {
		return dto.SpecOutput{}, fmt.Errorf("workflow execution failed: %s", err)
	}

	return result, nil
}

// TestConnection runs a workflow to test connection
func (t *Temporal) TestConnection(ctx context.Context, workflowID, flag, sourceType, version, config string) (map[string]interface{}, error) {
	params := &ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		WorkflowID: workflowID,
		Command:    docker.Check,
		Flag:       flag,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: TaskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, TestConnectionWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test connection workflow: %s", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %s", err)
	}

	return result, nil
}

// ManageSync handles all sync operations (create, update, delete, trigger)
func (t *Temporal) ManageSync(ctx context.Context, projectID string, jobID int, frequency string, action SyncAction) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("sync-%s-%d", projectID, jobID)
	scheduleID := fmt.Sprintf("schedule-%s", workflowID)

	handle := t.Client.ScheduleClient().GetHandle(ctx, scheduleID)
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
		return t.createSchedule(ctx, handle, scheduleID, workflowID, frequency, jobID)

	case ActionUpdate:
		if frequency == "" {
			return nil, fmt.Errorf("frequency is required for updating schedule")
		}
		return t.updateSchedule(ctx, handle, currentSchedule, scheduleID, frequency)

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
func (t *Temporal) createSchedule(ctx context.Context, _ client.ScheduleHandle, scheduleID, workflowID, cronSpec string, jobID int) (map[string]interface{}, error) {
	cronSpec = utils.ToCron(cronSpec)
	_, err := t.Client.ScheduleClient().Create(ctx, client.ScheduleOptions{
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
func (t *Temporal) updateSchedule(ctx context.Context, handle client.ScheduleHandle, currentSchedule *client.ScheduleDescription, _, cronSpec string) (map[string]interface{}, error) {
	cronSpec = utils.ToCron(cronSpec)
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

// cancelWorkflow cancels a workflow execution
func (t *Temporal) CancelWorkflow(ctx context.Context, workflowID, runID string) error {
	return t.Client.CancelWorkflow(ctx, workflowID, runID)
}

// ListWorkflow lists workflow executions based on the provided query
func (t *Temporal) ListWorkflow(ctx context.Context, request *workflowservice.ListWorkflowExecutionsRequest) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	// Query workflows using the SDK's ListWorkflow method
	resp, err := t.Client.ListWorkflow(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing workflow executions: %s", err)
	}

	return resp, nil
}
