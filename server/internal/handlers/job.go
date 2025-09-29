package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/utils"
)

type JobHandler struct {
	web.Controller
	jobService *services.JobService
}

func (c *JobHandler) Prepare() {
	var err error
	c.jobService, err = services.NewJobService()
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to initialize job service", err)
		return
	}
}

// @router /project/:projectid/jobs [get]
func (c *JobHandler) GetAllJobs() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobs, err := c.jobService.GetAllJobsByProject(projectID)
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
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.CreateJob(context.Background(), &req, projectID, userID); err != nil {
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
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.UpdateJob(context.Background(), &req, projectID, jobID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	id := GetIDFromPath(&c.Controller)
	jobName, err := c.jobService.DeleteJob(context.Background(), id)
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
	result, err := c.jobService.SyncJob(context.Background(), projectID, id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to sync job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /project/:projectid/jobs/:id/activate [put]
func (c *JobHandler) ActivateJob() {
	id := GetIDFromPath(&c.Controller)
	var req struct {
		Activate bool `json:"activate"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.ActivateJob(id, req.Activate, userID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		respondWithError(&c.Controller, statusCode, "Failed to activate job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, nil)
}

// @router /project/:projectid/jobs/:id/tasks [get]
func (c *JobHandler) GetJobTasks() {
	projectID := c.Ctx.Input.Param(":projectid")
	id := GetIDFromPath(&c.Controller)
	tasks, err := c.jobService.GetJobTasks(context.Background(), projectID, id)
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
	// Parse request body
	var req dto.JobTaskRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logs, err := c.jobService.GetTaskLogs(context.Background(), id, req.FilePath)
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

// @router /internal/worker/callback/presync/:id [get]
func (c *JobHandler) GetJobDetails() {
	jobId := GetIDFromPath(&c.Controller)

	if jobId == 0 {
		respondWithError(&c.Controller, http.StatusBadRequest, "job_id is required", nil)
		return
	}

	job, err := c.jobService.GetJobDetails(context.Background(), jobId)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get job details", err)
		return
	}

	utils.SuccessResponse(&c.Controller, job)
}

// @router /internal/worker/callback/postsync [post]
func (c *JobHandler) UpdateJobPostSync() {
	var req struct {
		JobID int    `json:"job_id"`
		State string `json:"state"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if req.JobID == 0 {
		respondWithError(&c.Controller, http.StatusBadRequest, "job_id is required", nil)
		return
	}
	if req.State == "" {
		respondWithError(&c.Controller, http.StatusBadRequest, "state is required", nil)
		return
	}

	if err := c.jobService.UpdateJobPostSync(context.Background(), req.JobID, req.State); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update job post sync", err)
		return
	}

	utils.SuccessResponse(&c.Controller, nil)
}

// @router /internal/worker/callback/sync-telemetry [post]
func (c *JobHandler) UpdateSyncTelemetry() {
	var req struct {
		JobID      int    `json:"job_id"`
		WorkflowID string `json:"workflow_id"`
		Event      string `json:"event"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
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
