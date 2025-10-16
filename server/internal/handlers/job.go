package handlers

import (
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/utils"
)

type JobHandler struct {
	web.Controller
}

// @router /project/:projectid/jobs [get]
func (c *JobHandler) GetAllJobs() {
	projectID := c.Ctx.Input.Param(":projectid")
	logger.Info("Get all jobs initiated - project_id=%s", projectID)

	jobs, err := svc.Job.GetAllJobs(projectID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs by project ID", err)
		return
	}
	utils.SuccessResponse(&c.Controller, jobs)
}

// @router /project/:projectid/jobs [post]
func (c *JobHandler) CreateJob() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.CreateJobRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Create job initiated - project_id=%s job_name=%s user_id=%v",
		projectID, req.Name, userID)

	if err := svc.Job.CreateJob(c.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to create job", err)
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
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Update job initiated - project_id=%s job_id=%d job_name=%s user_id=%v",
		projectID, jobID, req.Name, userID)

	if err := svc.Job.UpdateJob(c.Ctx.Request.Context(), &req, projectID, jobID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	id := GetIDFromPath(&c.Controller)
	logger.Info("Delete job initiated - job_id=%d", id)

	jobName, err := svc.Job.DeleteJob(c.Ctx.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, jobName)
}

// @router /project/:projectid/jobs/check-unique [post]
func (c *JobHandler) CheckUniqueJobName() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.CheckUniqueJobNameRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Check unique job name initiated - project_id=%s job_name=%s", projectID, req.JobName)

	unique, err := svc.Job.IsJobNameUnique(c.Ctx.Request.Context(), projectID, req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to check job name uniqueness", err)
		return
	}
	utils.SuccessResponse(&c.Controller, dto.CheckUniqueJobNameResponse{Unique: unique})
}

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	logger.Info("Sync job initiated - project_id=%s job_id=%d", projectID, id)

	result, err := svc.Job.SyncJob(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to sync job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /project/:projectid/jobs/:id/activate [put]
func (c *JobHandler) ActivateJob() {
	id := GetIDFromPath(&c.Controller)

	var req dto.JobStatusRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Activate job initiated - job_id=%d user_id=%v", id, userID)

	if err := svc.Job.ActivateJob(c.Ctx.Request.Context(), id, req, userID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, "Failed to activate job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, nil)
}

// @router /project/:projectid/jobs/:id/cancel [post]
func (c *JobHandler) CancelJobRun() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	logger.Info("Cancel job run initiated - project_id=%s job_id=%d", projectID, id)

	resp, err := svc.Job.CancelJobRun(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to cancel job run", err)
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/jobs/:id/tasks [get]
func (c *JobHandler) GetJobTasks() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	logger.Info("Get job tasks initiated - project_id=%s job_id=%d", projectID, id)

	tasks, err := svc.Job.GetJobTasks(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, "Failed to get job tasks", err)
		return
	}
	utils.SuccessResponse(&c.Controller, tasks)
}

// @router /project/:projectid/jobs/:id/logs [get]
func (c *JobHandler) GetTaskLogs() {
	id := GetIDFromPath(&c.Controller)

	var req dto.JobTaskRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Get task logs initiated - job_id=%d file_path=%s", id, req.FilePath)

	logs, err := svc.Job.GetTaskLogs(c.Ctx.Request.Context(), id, req.FilePath)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, "Failed to get task logs", err)
		return
	}
	utils.SuccessResponse(&c.Controller, logs)
}
