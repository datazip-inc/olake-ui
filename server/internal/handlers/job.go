package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/utils"
)

type JobHandler struct {
	web.Controller
	jobService *services.JobService
}

func (c *JobHandler) Prepare() {
	svc, err := services.NewJobService()
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to initialize job service", err)
		return
	}
	c.jobService = svc
}

// @router /project/:projectid/jobs [get]
func (c *JobHandler) GetAllJobs() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobs, err := c.jobService.GetAllJobs(projectID)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs by project ID", err)
		return
	}
	utils.SuccessResponse(&c.Controller, jobs)
}

// @router /project/:projectid/jobs [post]
func (c *JobHandler) CreateJob() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.CreateJobRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)

	if err := c.jobService.CreateJob(c.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
}

// @router /project/:projectid/jobs/:id [put]
func (c *JobHandler) UpdateJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobID := GetIDFromPath(&c.Controller)

	var req dto.UpdateJobRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.UpdateJob(c.Ctx.Request.Context(), &req, projectID, jobID, userID); err != nil {
		if errors.Is(err, constants.ErrInProgress) {
			respondWithError(&c.Controller, http.StatusConflict, "Clear destination workflow is already in progress", err)
			return
		}
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	id := GetIDFromPath(&c.Controller)
	jobName, err := c.jobService.DeleteJob(c.Ctx.Request.Context(), id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, jobName)
}

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	result, err := c.jobService.SyncJob(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to sync job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /project/:projectid/jobs/:id/activate [put]
func (c *JobHandler) ActivateJob() {
	id := GetIDFromPath(&c.Controller)

	var req dto.JobStatusRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.ActivateJob(c.Ctx.Request.Context(), id, req.Activate, userID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		respondWithError(&c.Controller, statusCode, "Failed to activate job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, nil)
}

// @router /project/:projectid/jobs/:id/cancel [post]
func (c *JobHandler) CancelJobRun() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	resp, err := c.jobService.CancelJobRun(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to cancel job run", err)
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/jobs/:id/tasks [get]
func (c *JobHandler) GetJobTasks() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	tasks, err := c.jobService.GetJobTasks(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		respondWithError(&c.Controller, statusCode, "Failed to get job tasks", err)
		return
	}
	utils.SuccessResponse(&c.Controller, tasks)
}

// @router /project/:projectid/jobs/:id/logs [get]
func (c *JobHandler) GetTaskLogs() {
	id := GetIDFromPath(&c.Controller)
	var req dto.JobTaskRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logs, err := c.jobService.GetTaskLogs(c.Ctx.Request.Context(), id, req.FilePath)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		respondWithError(&c.Controller, statusCode, "Failed to get task logs", err)
		return
	}
	utils.SuccessResponse(&c.Controller, logs)
}

// @router /project/:projectid/jobs/check-unique [post]
func (c *JobHandler) CheckUniqueJobName() {
	projectId := c.Ctx.Input.Param(":projectid")

	var req dto.CheckUniqueJobNameRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	unique, err := c.jobService.CheckUniqueJobName(projectId, req.JobName)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to check job uniqness", err)
		return
	}
	utils.SuccessResponse(&c.Controller, dto.CheckUniqueJobNameResponse{Unique: unique})
}

// @router /project/:projectid/jobs/:id/clear-destination [get]
func (c *JobHandler) GetClearDestinationStatus() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobID := GetIDFromPath(&c.Controller)

	running, err := c.jobService.GetClearDestinationStatus(c.Ctx.Request.Context(), projectID, jobID)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get clear destination status", err)
		return
	}
	utils.SuccessResponse(&c.Controller, dto.ClearDestinationStatusResposne{Running: running})
}

// @router /project/:projectid/jobs/:id/clear-destination [post]
func (c *JobHandler) ClearDestination() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)

	result, err := c.jobService.ClearDestination(c.Ctx.Request.Context(), projectID, id, "")
	if err != nil {
		if errors.Is(err, constants.ErrInProgress) {
			respondWithError(&c.Controller, http.StatusConflict, "operation already in progress:", err)
			return
		}
		respondWithError(&c.Controller, http.StatusInternalServerError, "failed to clear destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /project/:projectid/jobs/:id/stream-difference [post]
func (c *JobHandler) StreamsDifference() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobID := GetIDFromPath(&c.Controller)

	var req dto.StreamDifferenceRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, constants.ValidationInvalidRequestFormat, err)
		return
	}

	diffStreams, err := c.jobService.StreamsDifference(c.Ctx.Request.Context(), projectID, jobID, req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "stream difference command failed", err)
		return
	}

	utils.SuccessResponse(&c.Controller, dto.StreamDifferenceResponse{
		StreamDifference: diffStreams,
	})
}

// worker handler

// @router /internal/worker/callback/sync-telemetry [post]
func (c *JobHandler) UpdateSyncTelemetry() {
	var req struct {
		JobID      int    `json:"job_id"`
		WorkflowID string `json:"workflow_id"`
		Event      string `json:"event"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if req.JobID == 0 || req.WorkflowID == "" {
		respondWithError(&c.Controller, http.StatusBadRequest, "job_id and workflow_id are required", nil)
		return
	}

	if err := c.jobService.UpdateSyncTelemetry(context.Background(), req.JobID, req.WorkflowID, req.Event); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update sync telemetry", err)
		return
	}

	utils.SuccessResponse(&c.Controller, nil)

}
