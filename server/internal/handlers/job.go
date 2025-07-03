package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	jobID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

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
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, nil)
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Delete job using service
	jobName, err := c.jobService.DeleteJob(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, models.DeleteDestinationResponse{
		Name: jobName,
	})
}

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Sync job using service
	result, err := c.jobService.SyncJob(c.Ctx.Request.Context(), projectID, id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, result)
}

// @router /project/:projectid/jobs/:id/activate [put]
func (c *JobHandler) ActivateJob() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Parse request body
	var req struct {
		Activate bool `json:"activate"`
	}
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

	// Activate/Deactivate job using service
	if err := c.jobService.ActivateJob(id, req.Activate, userID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, nil)
}

// @router /project/:projectid/jobs/:id/tasks [get]
func (c *JobHandler) GetJobTasks() {
	projectID := c.Ctx.Input.Param(":projectid")
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Get job tasks using service
	tasks, err := c.jobService.GetJobTasks(projectID, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, tasks)
}

// @router /project/:projectid/jobs/:id/logs [get]
func (c *JobHandler) GetJobLogs() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	filePath := c.GetString("file")
	if filePath == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "File path is required")
		return
	}

	// Get job logs using service
	logs, err := c.jobService.GetTaskLogs(id, filePath)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&c.Controller, statusCode, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, logs)
}
