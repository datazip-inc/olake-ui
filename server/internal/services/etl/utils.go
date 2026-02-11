package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/utils"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"golang.org/x/mod/semver"
)

// JobLastRunInfo holds the latest run information for a job
type JobLastRunInfo struct {
	LastRunTime  string
	LastRunState string
	LastRunType  string
}

// fetchLatestJobRunsByJobIDs batches workflow queries for multiple jobs into a single/few temporal API calls
func fetchLatestJobRunsByJobIDs(ctx context.Context, tempClient *temporal.Temporal, projectID string, jobs []*models.Job) (map[int]JobLastRunInfo, error) {
	if len(jobs) == 0 {
		return map[int]JobLastRunInfo{}, nil
	}

	jobIDSet := make(map[int]struct{}, len(jobs))
	for _, job := range jobs {
		jobIDSet[job.ID] = struct{}{}
	}

	// Query latest workflow execution per job ID for this project.
	// Using BETWEEN with 'z' suffix.
	// 'z' sorts after all digits/hyphens in standard collation, ensuring we capture the full range.
	query := fmt.Sprintf("WorkflowId BETWEEN 'sync-%s-' AND 'sync-%s-z'", projectID, projectID)

	result := make(map[int]JobLastRunInfo, len(jobs))
	var nextPageToken []byte

	// Paginate through visibility results until we have the latest run for each job (or pages end).
	for {
		resp, err := tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Query:         query,
			PageSize:      int32(constants.DefaultListWorkflowPageSize),
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list workflows: %s", err)
		}

		// Walk each execution returned in this page.
		for _, execution := range resp.Executions {
			jobID, ok := utils.ExtractJobIDFromWorkflowID(execution.Execution.WorkflowId, projectID)
			if !ok {
				continue
			}
			// skip if id not found in jobIDSet
			if _, inSet := jobIDSet[jobID]; !inSet {
				continue
			}
			// skip if already exists
			if _, exists := result[jobID]; exists {
				continue
			}

			// add the latest run for this job
			opType := syncWorkflowOperationType(execution)
			result[jobID] = JobLastRunInfo{
				LastRunTime:  execution.StartTime.AsTime().Format(time.RFC3339),
				LastRunState: execution.Status.String(),
				LastRunType:  utils.Ternary(opType == temporal.Sync, "sync", "clear").(string),
			}

			// break if all jobs are populated.
			if len(result) == len(jobIDSet) {
				return result, nil
			}
		}

		// break if no more executions are available.
		if len(resp.NextPageToken) == 0 {
			break
		}
		nextPageToken = resp.NextPageToken
	}

	return result, nil
}

func cancelAllJobWorkflows(ctx context.Context, tempClient *temporal.Temporal, jobs []*models.Job, projectID string) error {
	if len(jobs) == 0 {
		return nil
	}

	// Build combined query
	var conditions []string
	for _, job := range jobs {
		conditions = append(conditions, fmt.Sprintf(
			"(WorkflowId BETWEEN 'sync-%s-%d-' AND 'sync-%s-%d-z' AND OperationType != '%s')",
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

func buildJobDataItems(jobs []*models.Job, lastRunByJobID map[int]JobLastRunInfo, contextType string) ([]dto.JobDataItem, error) {
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

		// Set workflow info from pre-fetched map
		if lastRun, ok := lastRunByJobID[job.ID]; ok {
			jobInfo.LastRunTime = lastRun.LastRunTime
			jobInfo.LastRunState = lastRun.LastRunState
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
		"WorkflowId BETWEEN 'sync-%s-%d-' AND 'sync-%s-%d-z' AND OperationType = '%s' AND ExecutionStatus = 'Running'",
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

// checks the version compatibility for clear-destination and stream difference operation
// supported in versions >= v0.3.0
func CheckClearDestinationCompatibility(sourceVersion string) error {
	if semver.Compare(sourceVersion, constants.DefaultClearDestinationVersion) < 0 {
		return fmt.Errorf("source version %s is not supported for clear destination. please update the source version to %s or higher", sourceVersion, constants.DefaultClearDestinationVersion)
	}
	return nil
}
