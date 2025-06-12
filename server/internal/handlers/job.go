package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/services"
	"github.com/datazip/olake-frontend/server/utils"
)

type JobHandler struct {
	web.Controller
	jobService *services.JobService
}

// Prepare initializes the service instances
func (c *JobHandler) Prepare() {
	var err error
	c.jobService, err = services.NewJobService()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to initialize job service")
		return
	}
}

// @router /project/:projectid/jobs [get]
func (c *JobHandler) GetAllJobs() {
	projectID := c.Ctx.Input.Param(":projectid")

	jobs, err := c.jobService.GetAllJobsByProject(projectID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs by project ID")
		return
	}

	utils.SuccessResponse(&c.Controller, jobs)
}

// @router /project/:projectid/jobs [post]
func (c *JobHandler) CreateJob() {
	projectID := c.Ctx.Input.Param(":projectid")

	// Parse request body
	var req models.CreateJobRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get user ID from session
	var userID *int
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if id, ok := sessionUserID.(int); ok {
			userID = &id
		}
	}

	// Create job using service
	if err := c.jobService.CreateJob(c.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/jobs/:id [put]
func (c *JobHandler) UpdateJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobID := GetIDFromPath(&c.Controller)

	// Parse request body
	var req models.UpdateJobRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get user ID from session
	var userID *int
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if id, ok := sessionUserID.(int); ok {
			userID = &id
		}
	}

	// Update job using service
	if err := c.jobService.UpdateJob(c.Ctx.Request.Context(), &req, projectID, jobID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	jobName, err := c.jobService.DeleteJob(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete job")
		return
	}

	utils.SuccessResponse(&c.Controller, models.DeleteDestinationResponse{
		Name: jobName,
	})
}

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	id := GetIDFromPath(&c.Controller)
	projectID := c.Ctx.Input.Param(":projectid")

	resp, err := c.jobService.SyncJob(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		if err.Error() == "job not found: record not found" || err.Error() == "job not found" {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
			return
		}
		if err.Error() == "job must have both source and destination configured" {
			utils.ErrorResponse(&c.Controller, http.StatusBadRequest, err.Error())
			return
		}
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/jobs/:id/activate [post]
func (c *JobHandler) ActivateJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Parse request body
	var req models.JobStatus
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get user ID from session
	var userID *int
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if id, ok := sessionUserID.(int); ok {
			userID = &id
		}
	}

	if err := c.jobService.ActivateJob(id, req.Activate, userID); err != nil {
		if err.Error() == "job not found: record not found" || err.Error() == "job not found" {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
			return
		}
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job activation status")
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/jobs/:id/tasks [get]
func (c *JobHandler) GetJobTasks() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	projectID := c.Ctx.Input.Param(":projectid")

	tasks, err := c.jobService.GetJobTasks(projectID, id)
	if err != nil {
		if err.Error() == "job not found: record not found" || err.Error() == "job not found" {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
			return
		}
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, tasks)
}

// @router /project/:projectid/jobs/:id/tasks/:taskid/logs [post]
func (c *JobHandler) GetTaskLogs() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Parse request body
	var req struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	logs, err := c.jobService.GetTaskLogs(id, req.FilePath)
	if err != nil {
		if err.Error() == "job not found: record not found" || err.Error() == "job not found" {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, err.Error())
			return
		}
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, logs)
}
