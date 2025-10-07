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
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.CreateJob(context.Background(), &req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)
	unique, err := c.jobORM.IsJobNameUnique(projectIDStr, req.Name)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to check job name uniqueness")
		return
	}
	if !unique {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Job name already exists")
		return
	}
	// Find or create source
	source, err := c.getOrCreateSource(&req.Source, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process source: %s", err))
		return
	}

	// Find or create destination
	dest, err := c.getOrCreateDestination(&req.Destination, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process destination: %s", err))
		return
	}

	// Create job model
	job := &models.Job{
		Name:          req.Name,
		SourceID:      source,
		DestID:        dest,
		Active:        true,
		Frequency:     req.Frequency,
		StreamsConfig: req.StreamsConfig,
		State:         "{}",
		ProjectID:     projectIDStr,
	}

	// Get user information from session
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		job.CreatedBy = user
		job.UpdatedBy = user
	}

	if err := c.jobORM.Create(job); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	// telemetry events
	telemetry.TrackJobCreation(context.Background(), job)

	if c.tempClient != nil {
		fmt.Println("Using Temporal workflow for sync job")
		_, err = c.tempClient.ManageSync(
			c.Ctx.Request.Context(),
			job.ProjectID,
			job.ID,
			job.Frequency,
			temporal.ActionCreate,
		)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Temporal workflow execution failed for create job schedule: %s", err))
		}
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/jobs/:id [put]
func (c *JobHandler) UpdateJob() {
	projectID := c.Ctx.Input.Param(":projectid")
	jobID := GetIDFromPath(&c.Controller)
	var req dto.UpdateJobRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.jobService.UpdateJob(context.Background(), &req, projectID, jobID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update job", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req.Name)

	// Find or create source
	source, err := c.getOrCreateSource(&req.Source, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process source: %s", err))
		return
	}

	// Find or create destination
	dest, err := c.getOrCreateDestination(&req.Destination, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process destination: %s", err))
		return
	}

	// Update fields
	existingJob.Name = req.Name
	existingJob.SourceID = source
	existingJob.DestID = dest
	existingJob.Active = req.Activate
	existingJob.Frequency = req.Frequency
	existingJob.StreamsConfig = req.StreamsConfig
	existingJob.UpdatedAt = time.Now()
	existingJob.ProjectID = projectIDStr

	// Update user information
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		existingJob.UpdatedBy = user
	}

	// cancel existing workflow
	err = cancelJobWorkflow(c.tempClient, existingJob, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to cancel workflow for job %s", err))
		return
	}
	// Update job in database
	if err := c.jobORM.Update(existingJob); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job")
		return
	}

	// Track sources and destinations status after job update
	telemetry.TrackJobEntity(context.Background())

	if c.tempClient != nil {
		logs.Info("Using Temporal workflow for sync job")
		_, err = c.tempClient.ManageSync(
			c.Ctx.Request.Context(),
			existingJob.ProjectID,
			existingJob.ID,
			existingJob.Frequency,
			temporal.ActionUpdate,
		)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Temporal workflow execution failed for update job schedule: %s", err))
			return
		}
	}

	utils.SuccessResponse(&c.Controller, req)
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

	// Get job name for response
	job, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}
	// cancel existing workflow
	err = cancelJobWorkflow(c.tempClient, job, job.ProjectID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to cancel workflow for job %s", err))
		return
	}
	jobName := job.Name
	if c.tempClient != nil {
		logs.Info("Using Temporal workflow for delete job schedule")
		_, err = c.tempClient.ManageSync(
			c.Ctx.Request.Context(),
			job.ProjectID,
			job.ID,
			job.Frequency,
			temporal.ActionDelete,
		)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Temporal workflow execution failed for delete job schedule: %s", err))
			return
		}
	}

	// Delete job
	if err := c.jobORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete job")
		return
	}

	// Track sources and destinations status after job deletion
	telemetry.TrackJobEntity(context.Background())

	utils.SuccessResponse(&c.Controller, models.DeleteDestinationResponse{
		Name: jobName,
	})
}

// @router /project/:projectid/jobs/check-unique [post]
func (c *JobHandler) CheckUniqueJobName() {
	projectIDStr := c.Ctx.Input.Param(":projectid")
	var req models.CheckUniqueJobNameRequest

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	if req.JobName == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Job name is required")
		return
	}
	unique, err := c.jobORM.IsJobNameUnique(projectIDStr, req.JobName)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to check job name uniqueness")
		return
	}
	utils.SuccessResponse(&c.Controller, models.CheckUniqueJobNameResponse{
		Unique: unique,
	})
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
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
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

// @router /project/:projectid/jobs/:id/activate [post]
func (c *JobHandler) ActivateJob() {
	id := GetIDFromPath(&c.Controller)

	// Parse request body
	var req models.JobStatus
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get existing job
	job, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}
	action := temporal.ActionUnpause
	if !req.Activate {
		action = temporal.ActionPause
	}
	if c.tempClient != nil {
		logs.Info("Using Temporal workflow for activate job schedule")
		_, err = c.tempClient.ManageSync(
			c.Ctx.Request.Context(),
			job.ProjectID,
			job.ID,
			job.Frequency,
			action,
		)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Temporal workflow execution failed for activate job schedule: %s", err))
			return
		}
	}
	// Update activation status
	job.Active = req.Activate
	job.UpdatedAt = time.Now()

	// Update user information
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		job.UpdatedBy = user
	}

	// Update job in database
	if err := c.jobORM.Update(job); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job activation status")
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/jobs/:id/cancel [get]
func (c *JobHandler) CancelJobRun() {
	// Parse inputs
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}
	projectID := c.Ctx.Input.Param(":projectid")

	// Ensure job exists
	job, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, fmt.Sprintf("Job not found: %v", err))
		return
	}

	if err := cancelJobWorkflow(c.tempClient, job, projectID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("job workflow cancel failed: %v", err))
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]any{
		"message": "job workflow cancel requested successfully",
	})
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
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
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
