package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/internal/temporal"
	"go.temporal.io/api/workflowservice/v1"
)

func cancelAllJobWorkflows(ctx context.Context, tempClient *temporal.Temporal, jobs []*models.Job, projectID string) error {
	if len(jobs) == 0 {
		return nil
	}

	// Build combined query
	var conditions []string
	for _, job := range jobs {
		conditions = append(conditions, fmt.Sprintf(
			"(WorkflowId BETWEEN 'sync-%s-%d' AND 'sync-%s-%d-~')",
			projectID, job.ID, projectID, job.ID,
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

func buildJobDataItems(jobs []*models.Job) ([]dto.JobDataItem, error) {
	jobItems := make([]dto.JobDataItem, 0, len(jobs))
	for _, job := range jobs {
		jobItems = append(jobItems, dto.JobDataItem{ID: job.ID, Name: job.Name})
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
