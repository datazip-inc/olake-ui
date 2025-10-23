package services

import (
	"context"
	"fmt"

	"github.com/datazip/olake-ui/server/internal/temporal"
	"go.temporal.io/api/workflowservice/v1"
)

func cancelJobWorkflow(ctx context.Context, tempClient *temporal.Client, projectID string, jobID int) error {
	query := fmt.Sprintf(
		"WorkflowId BETWEEN 'sync-%s-%d' AND 'sync-%s-%d-~' AND ExecutionStatus = 'Running'",
		projectID, jobID, projectID, jobID,
	)

	resp, err := tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return fmt.Errorf("list workflows failed: %s", err)
	}
	if len(resp.Executions) == 0 {
		return nil // no running workflows
	}

	for _, wfExec := range resp.Executions {
		if err := tempClient.CancelWorkflow(ctx,
			wfExec.Execution.WorkflowId, wfExec.Execution.RunId); err != nil {
			return fmt.Errorf("failed to cancel workflow[%s]: %s", wfExec.Execution.WorkflowId, err)
		}
	}
	return nil
}

func isClearRunning(ctx context.Context, tempClient *temporal.Client, projectID string, jobID int) (bool, error) {
	query := fmt.Sprintf("WorkflowId BETWEEN 'clear-destination-%s-%d' AND 'clear-destination-%s-%d-~' AND ExecutionStatus = 'Running'",
		projectID, jobID, projectID, jobID,
	)

	resp, err := tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return false, err
	}
	return len(resp.Executions) > 0, nil
}

func isSyncRunning(ctx context.Context, tempClient *temporal.Client, projectID string, jobID int) (bool, error) {
	query := fmt.Sprintf(
		"WorkflowId BETWEEN 'sync-%s-%d' AND 'sync-%s-%d-~' AND ExecutionStatus = 'Running'",
		projectID, jobID, projectID, jobID,
	)

	resp, err := tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return false, err
	}
	return len(resp.Executions) > 0, nil
}

// check if stream difference exists in the diff catalog
func streamDifferenceExists(diffCatalog map[string]interface{}) bool {
	if diffCatalog == nil {
		return false
	}
	if streams, ok := diffCatalog["streams"].([]interface{}); ok {
		return len(streams) > 0
	}
	return false
}
