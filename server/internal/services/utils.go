package services

import (
	"context"
	"fmt"

	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/temporal"
	"go.temporal.io/api/workflowservice/v1"
)

func cancelJobWorkflow(tempClient *temporal.Client, job *models.Job, projectID string) error {
	query := fmt.Sprintf(
		"WorkflowId BETWEEN 'sync-%s-%d' AND 'sync-%s-%d-~' AND ExecutionStatus = 'Running'",
		projectID, job.ID, projectID, job.ID,
	)

	resp, err := tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return fmt.Errorf("list workflows failed: %s", err)
	}
	if len(resp.Executions) == 0 {
		return nil // no running workflows
	}

	for _, wfExec := range resp.Executions {
		if err := tempClient.CancelWorkflow(context.Background(),
			wfExec.Execution.WorkflowId, wfExec.Execution.RunId); err != nil {
			return fmt.Errorf("failed to cancel workflow[%s]: %s", wfExec.Execution.WorkflowId, err)
		}
	}
	return nil
}
