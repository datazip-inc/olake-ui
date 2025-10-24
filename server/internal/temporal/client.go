package temporal

import (
	"context"
	"fmt"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/utils"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type Temporal struct {
	Client    client.Client
	taskQueue string
}

// NewClient creates a new Temporal client
func NewClient() (*Temporal, error) {
	// Choose task queue based on deployment mode
	temporalAddress := web.AppConfig.DefaultString("TEMPORAL_ADDRESS", "localhost:7233")
	taskQueue := constants.DockerTaskQueue
	if web.AppConfig.DefaultString("DEPLOYMENT_MODE", "docker") == "kubernetes" {
		taskQueue = constants.K8sTaskQueue
	}

	c, err := client.Dial(client.Options{
		HostPort: temporalAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %s", err)
	}

	return &Temporal{
		Client:    c,
		taskQueue: taskQueue,
	}, nil
}

// Close closes the Temporal client
func (t *Temporal) Close() {
	if t.Client != nil {
		t.Client.Close()
	}
}

func (t *Temporal) WorkflowAndScheduleID(projectID string, jobID int) (string, string) {
	workflowID := fmt.Sprintf("sync-%s-%d", projectID, jobID)
	return workflowID, fmt.Sprintf("schedule-%s", workflowID)
}

// createSchedule creates a new schedule
func (t *Temporal) CreateSchedule(ctx context.Context, frequency, projectID string, jobID int) error {
	workflowID, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)
	cronExpression := utils.ToCron(frequency)
	_, err := t.Client.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{cronExpression},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			Workflow:  RunSyncWorkflow,
			Args:      []any{jobID},
			TaskQueue: t.taskQueue,
		},
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
	})
	return err
}

// updateSchedule updates an existing schedule
func (t *Temporal) UpdateSchedule(ctx context.Context, frequency, projectID string, jobID int) error {
	cronExpression := utils.ToCron(frequency)
	_, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)

	handle := t.Client.ScheduleClient().GetHandle(ctx, scheduleID)
	return handle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			input.Description.Schedule.Spec = &client.ScheduleSpec{
				CronExpressions: []string{cronExpression},
			}
			return &client.ScheduleUpdate{
				Schedule: &input.Description.Schedule,
			}, nil
		},
	})
}

func (t *Temporal) PauseSchedule(ctx context.Context, projectID string, jobID int) error {
	_, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)
	return t.Client.ScheduleClient().GetHandle(ctx, scheduleID).Pause(ctx, client.SchedulePauseOptions{
		Note: "user paused the schedule",
	})
}

func (t *Temporal) UnpauseSchedule(ctx context.Context, projectID string, jobID int) error {
	_, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)
	return t.Client.ScheduleClient().GetHandle(ctx, scheduleID).Unpause(ctx, client.ScheduleUnpauseOptions{
		Note: "user paused the schedule",
	})
}

func (t *Temporal) DeleteSchedule(ctx context.Context, projectID string, jobID int) error {
	_, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)
	return t.Client.ScheduleClient().GetHandle(ctx, scheduleID).Delete(ctx)
}

func (t *Temporal) TriggerSchedule(ctx context.Context, projectID string, jobID int) error {
	_, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)
	return t.Client.ScheduleClient().GetHandle(ctx, scheduleID).Trigger(ctx, client.ScheduleTriggerOptions{
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
	})
}
