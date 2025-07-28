package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/utils"
)

// get id from path
func GetIDFromPath(c *web.Controller) int {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id")
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

// Helper to bind and validate JSON request
func bindJSON(c *web.Controller, target interface{}) error {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, target); err != nil {
		return fmt.Errorf("invalid request format: %s", err)
	}
	return nil
}

// // buildJobDataItems creates job data items with workflow information
// // Returns (jobItems, success). If success is false, an error occurred and the handler should return.
// func buildJobDataItems(jobs []*models.Job, err error, projectIDStr, contextType string, tempClient *temporal.Client, controller *web.Controller) ([]models.JobDataItem, bool) {
// 	jobItems := make([]models.JobDataItem, 0)

// 	if err != nil {
// 		return jobItems, true // No jobs is OK, return empty slice
// 	}

// 	for _, job := range jobs {
// 		jobInfo := models.JobDataItem{
// 			Name:     job.Name,
// 			ID:       job.ID,
// 			Activate: job.Active,
// 		}

// 		// Set source/destination info based on context
// 		if contextType == "source" && job.DestID != nil {
// 			jobInfo.DestinationName = job.DestID.Name
// 			jobInfo.DestinationType = job.DestID.DestType
// 		} else if contextType == "destination" && job.SourceID != nil {
// 			jobInfo.SourceName = job.SourceID.Name
// 			jobInfo.SourceType = job.SourceID.Type
// 		}

// 		if !setJobWorkflowInfo(&jobInfo, job.ID, projectIDStr, tempClient, controller) {
// 			return nil, false // Error occurred, signal failure
// 		}
// 		jobItems = append(jobItems, jobInfo)
// 	}

// 	return jobItems, true
// }

// // setJobWorkflowInfo fetches and sets workflow execution information for a job
// // Returns false if an error occurred that should stop processing
// func setJobWorkflowInfo(jobInfo *models.JobDataItem, jobID int, projectIDStr string, tempClient *temporal.Client, controller *web.Controller) bool {
// 	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectIDStr, jobID, projectIDStr, jobID)

// 	resp, err := tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
// 		Query:    query,
// 		PageSize: 1,
// 	})

// 	if err != nil {
// 		utils.ErrorResponse(controller, http.StatusInternalServerError, fmt.Sprintf("failed to list workflows: %v", err))
// 		return false
// 	}

// 	if len(resp.Executions) > 0 {
// 		jobInfo.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
// 		jobInfo.LastRunState = resp.Executions[0].Status.String()
// 	} else {
// 		jobInfo.LastRunTime = ""
// 		jobInfo.LastRunState = ""
// 	}
// 	return true
// }
