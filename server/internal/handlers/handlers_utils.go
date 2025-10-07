package handlers

import (
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/utils"
)

// get id from path
func GetIDFromPath(c *web.Controller) int {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid id", err)
		return 0
	}
	return id
}

// Helper to log and respond with error
func respondWithError(c *web.Controller, statusCode int, msg string, err error) {
	if err != nil {
		logs.Error("%s: %s", msg, err)
	}
	utils.ErrorResponse(c, statusCode, msg)
}

// Helper to extract user ID from session
func GetUserIDFromSession(c *web.Controller) *int {
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if uid, ok := sessionUserID.(int); ok {
			return &uid
		}
	}
	return nil
}

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
