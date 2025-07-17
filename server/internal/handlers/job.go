package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/services"
	"github.com/datazip/olake-frontend/server/utils"
)

type JobHandler struct {
	web.Controller
	jobService *services.JobService
}

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
	var req models.CreateJobRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.CreateJob(context.Background(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
}

// @router /project/:projectid/jobs/:id [put]
func (c *JobHandler) UpdateJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}
	var req models.UpdateJobRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.UpdateJob(context.Background(), &req, projectID, jobID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}
	jobName, err := c.jobService.DeleteJob(context.Background(), id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(&c.Controller, jobName)
}

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}
	result, err := c.jobService.SyncJob(context.Background(), projectID, id)
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
	var req struct {
		Activate bool `json:"activate"`
	}
	if err := bindJSON(&c.Controller, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
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
	tasks, err := c.jobService.GetJobTasks(context.Background(), projectID, id)
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
	if err := bindJSON(&c.Controller, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	logs, err := c.jobService.GetTaskLogs(context.Background(), id, req.FilePath)
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
