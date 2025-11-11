package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
)

func cancelAllJobWorkflows(ctx context.Context, tempClient *temporal.Temporal, jobs []*models.Job, projectID string) error {
	if len(jobs) == 0 {
		return nil
	}

	// Build combined query
	var conditions []string
	for _, job := range jobs {
		conditions = append(conditions, fmt.Sprintf(
			"(WorkflowId BETWEEN 'sync-%s-%d' AND 'sync-%s-%d-~' AND OperationType != '%s')",
			projectID, job.ID, projectID, job.ID, temporal.ClearDestination,
		))
	}

	query := fmt.Sprintf("(%s) AND ExecutionStatus = 'Running'", strings.Join(conditions, " OR "))

	// List all running workflows at once
	resp, err := tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return fmt.Errorf("list workflows failed: %s", err)
	}
	if len(resp.Executions) == 0 {
		return nil // no running workflows
	}

	// Cancel each found workflow (still a loop, but only one list RPC)
	for _, wfExec := range resp.Executions {
		if err := tempClient.CancelWorkflow(ctx,
			wfExec.Execution.WorkflowId, wfExec.Execution.RunId); err != nil {
			return fmt.Errorf("failed to cancel workflow[%s]: %s", wfExec.Execution.WorkflowId, err)
		}
	}
	return nil
}

func buildJobDataItems(jobs []*models.Job, tempClient *temporal.Temporal, contextType string) ([]dto.JobDataItem, error) {
	jobItems := make([]dto.JobDataItem, 0)
	for _, job := range jobs {
		jobInfo := dto.JobDataItem{
			Name:     job.Name,
			ID:       job.ID,
			Activate: job.Active,
		}

		// Set source/destination info based on context
		if contextType == "source" && job.DestID != nil {
			jobInfo.DestinationName = job.DestID.Name
			jobInfo.DestinationType = job.DestID.DestType
		} else if contextType == "destination" && job.SourceID != nil {
			jobInfo.SourceName = job.SourceID.Name
			jobInfo.SourceType = job.SourceID.Type
		}

		if err := setJobWorkflowInfo(&jobInfo, job.ID, job.ProjectID, tempClient); err != nil {
			return nil, fmt.Errorf("failed to set job workflow info: %s", err)
		}
		jobItems = append(jobItems, jobInfo)
	}

	return jobItems, nil
}

func setUsernames(createdBy, updatedBy *string, createdByUser, updatedByUser *models.User) {
	if createdByUser != nil {
		*createdBy = createdByUser.Username
	}
	if updatedByUser != nil {
		*updatedBy = updatedByUser.Username
	}
}

// setJobWorkflowInfo fetches and sets workflow execution information for a job
// Returns false if an error occurred that should stop processing
func setJobWorkflowInfo(jobInfo *dto.JobDataItem, jobID int, projectID string, tempClient *temporal.Temporal) error {
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, jobID, projectID, jobID)

	resp, err := tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query:    query,
		PageSize: 1,
	})

	if err != nil {
		return fmt.Errorf("failed to list workflows: %s", err)
	}

	if len(resp.Executions) > 0 {
		jobInfo.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
		jobInfo.LastRunState = resp.Executions[0].Status.String()
	} else {
		jobInfo.LastRunTime = ""
		jobInfo.LastRunState = ""
	}
	return nil
}

// Checks if the sync worklfow run is "sync" or "clear-destination"
func syncWorkflowOperationType(execution *workflow.WorkflowExecutionInfo) temporal.Command {
	if execution.SearchAttributes == nil {
		return temporal.Sync
	}

	opTypePayload, ok := execution.SearchAttributes.IndexedFields["OperationType"]
	if !ok || opTypePayload == nil {
		return temporal.Sync
	}

	var opType string
	if err := converter.GetDefaultDataConverter().FromPayload(opTypePayload, &opType); err == nil {
		return temporal.Command(opType)
	}

	return temporal.Sync
}

// isWorkflowRunning checks if workflows of a specific type are running
func isWorkflowRunning(ctx context.Context, tempClient *temporal.Temporal, projectID string, jobID int, opType temporal.Command) (bool, []*workflow.WorkflowExecutionInfo, error) {
	query := fmt.Sprintf(
		"WorkflowId BETWEEN 'sync-%s-%d' AND 'sync-%s-%d-~' AND OperationType = '%s' AND ExecutionStatus = 'Running'",
		projectID, jobID, projectID, jobID, opType,
	)

	resp, err := tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query:    query,
		PageSize: 1,
	})
	if err != nil {
		return false, nil, err
	}
	return len(resp.Executions) > 0, resp.Executions, nil
}

// waitForSyncToStop checks if a sync workflow is running and optionally waits for it to stop.
// - If sync is not running: returns nil immediately
// - If sync is running and maxWaitTime <= 0: returns error immediately (no wait)
// - If sync is running and maxWaitTime > 0: waits up to maxWaitTime for sync to complete
func waitForSyncToStop(ctx context.Context, tempClient *temporal.Temporal, projectID string, jobID int, maxWaitTime time.Duration) error {
	isSyncRunning, executions, err := isWorkflowRunning(ctx, tempClient, projectID, jobID, temporal.Sync)
	if err != nil {
		return fmt.Errorf("failed to check sync status: %s", err)
	}
	if !isSyncRunning {
		return nil
	}

	if maxWaitTime <= 0 {
		return fmt.Errorf("sync is in progress, please wait or cancel the sync")
	}

	timedCtx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	workflowID := executions[0].Execution.WorkflowId
	runID := executions[0].Execution.RunId

	_, err = tempClient.Client.WorkflowService().GetWorkflowExecutionHistory(
		timedCtx,
		&workflowservice.GetWorkflowExecutionHistoryRequest{
			Namespace: "default",
			Execution: &commonpb.WorkflowExecution{
				WorkflowId: workflowID,
				RunId:      runID,
			},
			WaitNewEvent:           true,
			HistoryEventFilterType: enumspb.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT,
		},
	)
	if err != nil || timedCtx.Err() != nil {
		return fmt.Errorf("timeout waiting for sync to stop after %v", maxWaitTime)
	}

	return nil
}
