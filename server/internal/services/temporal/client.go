package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

type Temporal struct {
	Client    client.Client
	taskQueue string
}

// NewClient creates a new Temporal client
func NewClient() (*Temporal, error) {
	temporalAddress, err := web.AppConfig.String(constants.ConfTemporalAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get temporal address: %s", err)
	}

	logger.Info("TEMPORAL INIT DONE", temporalAddress)

	var temporalClient *Temporal
	err = utils.RetryWithBackoff(func() error {
		client, dialErr := client.Dial(client.Options{
			HostPort: temporalAddress,
		})
		if dialErr != nil {
			return fmt.Errorf("failed to create temporal client: %s", dialErr)
		}

		temporalClient = &Temporal{
			Client:    client,
			taskQueue: constants.TemporalTaskQueue,
		}
		return nil
	}, 3, time.Second)
	if err != nil {
		return nil, err
	}

	return temporalClient, nil
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
func (t *Temporal) CreateSchedule(ctx context.Context, job *models.Job) error {
	workflowID, scheduleID := t.WorkflowAndScheduleID(job.ProjectID, job.ID)
	cronExpression := utils.ToCron(job.Frequency)

	req := buildExecutionReqForSync(job, workflowID)

	_, err := t.Client.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{cronExpression},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			Workflow:  RunSyncWorkflow,
			Args:      []any{req},
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

func (t *Temporal) ResumeSchedule(ctx context.Context, projectID string, jobID int) error {
	_, scheduleID := t.WorkflowAndScheduleID(projectID, jobID)
	return t.Client.ScheduleClient().GetHandle(ctx, scheduleID).Unpause(ctx, client.ScheduleUnpauseOptions{
		Note: "user resumed the schedule",
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
