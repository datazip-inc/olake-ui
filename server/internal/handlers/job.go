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
	// Find or create source
	source, err := c.getOrCreateSource(req.Source, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process source: %s", err))
		return
	}

	// Find or create destination
	dest, err := c.getOrCreateDestination(req.Destination, projectIDStr)
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
	// Set user information
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		job.CreatedBy = user
		job.UpdatedBy = user
	}

	// Create job in database
	if err := c.jobORM.Create(job); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

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
			fmt.Printf("Temporal workflow execution failed: %v", err)
		} else {
			fmt.Println("Successfully executed sync job via Temporal")
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
	// Get existing job
	existingJob, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}

	// Find or create source
	source, err := c.getOrCreateSource(req.Source, projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process source: %s", err))
		return
	}

	// Find or create destination
	dest, err := c.getOrCreateDestination(req.Destination, projectIDStr)
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

	// Update job in database
	if err := c.jobORM.Update(existingJob); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job")
		return
	}
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
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Temporal workflow execution failed: %s", err))
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
	// Get job name for response
	job, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
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
		}
	}
	// Delete job
	if err := c.jobORM.Delete(id); err != nil {
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
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}
	// Check if job exists
	job, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}

	// Validate source and destination exist
	if job.SourceID == nil || job.DestID == nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Job must have both source and destination configured")
		return
	}

	if c.tempClient != nil {
		logs.Info("Using Temporal workflow for sync job")
		_, err = c.tempClient.ManageSync(
			c.Ctx.Request.Context(),
			job.ProjectID,
			job.ID,
			job.Frequency,
			temporal.ActionTrigger,
		)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Temporal workflow execution failed: %s", err))
		}
	}

	utils.SuccessResponse(&c.Controller, resp)
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
	// Get job to verify it exists
	job, err := c.jobORM.GetByID(id, true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}
	var tasks []models.JobTask
	// Construct a query for workflows related to this project and job
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectIDStr, job.ID, projectIDStr, job.ID)
	// List workflows using the direct query
	resp, err := c.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to list workflows: %v", err))
		return
	}
	for _, execution := range resp.Executions {
		startTime := execution.StartTime.AsTime().UTC()
		var runTime string
		if execution.CloseTime != nil {
			runTime = execution.CloseTime.AsTime().UTC().Sub(startTime).Round(time.Second).String()
		} else {
			runTime = time.Since(startTime).Round(time.Second).String()
		}
		tasks = append(tasks, models.JobTask{
			Runtime:   runTime,
			StartTime: startTime.Format(time.RFC3339),
			Status:    execution.Status.String(),
			FilePath:  execution.Execution.WorkflowId,
		})
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
	// Verify job exists
	_, err = c.jobORM.GetByID(id, true)
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

		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}
		if logEntry.Level != "debug" {
			logs = append(logs, map[string]interface{}{
				"level":   logEntry.Level,
				"time":    logEntry.Time.UTC().Format(time.RFC3339),
				"message": logEntry.Message,
			})
		}
	}

	utils.SuccessResponse(&c.Controller, logs)
}

// Helper methods

// getOrCreateSource finds or creates a source based on the provided config
func (c *JobHandler) getOrCreateSource(config models.JobSourceConfig, projectIDStr string) (*models.Source, error) {
	// Try to find an existing source matching the criteria
	sources, err := c.sourceORM.GetByNameAndType(config.Name, config.Type, projectIDStr)
	if err == nil && len(sources) > 0 {
		// Update the existing source if found
		source := sources[0]
		source.Config = config.Config
		source.Version = config.Version

		// Get user info for update
		userID := c.GetSession(constants.SessionUserID)
		if userID != nil {
			user := &models.User{ID: userID.(int)}
			source.UpdatedBy = user
		}

		if err := c.sourceORM.Update(source); err != nil {
			return nil, fmt.Errorf("failed to update source: %s", err)
		}

		return source, nil
	}

	// Create a new source if not found
	source := &models.Source{
		Name:      config.Name,
		Type:      config.Type,
		Config:    config.Config,
		Version:   config.Version,
		ProjectID: projectIDStr,
	}

	// Set user info
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		source.CreatedBy = user
		source.UpdatedBy = user
	}

	if err := c.sourceORM.Create(source); err != nil {
		return nil, fmt.Errorf("failed to create source: %s", err)
	}

	return source, nil
}

// getOrCreateDestination finds or creates a destination based on the provided config
func (c *JobHandler) getOrCreateDestination(config models.JobDestinationConfig, projectIDStr string) (*models.Destination, error) {
	// Try to find an existing destination matching the criteria
	destinations, err := c.destORM.GetByNameAndType(config.Name, config.Type, projectIDStr)
	if err == nil && len(destinations) > 0 {
		// Update the existing destination if found
		dest := destinations[0]
		dest.Config = config.Config
		dest.Version = config.Version

		// Get user info for update
		userID := c.GetSession(constants.SessionUserID)
		if userID != nil {
			user := &models.User{ID: userID.(int)}
			dest.UpdatedBy = user
		}

		if err := c.destORM.Update(dest); err != nil {
			return nil, fmt.Errorf("failed to update destination: %s", err)
		}

		return dest, nil
	}

	// Create a new destination if not found
	dest := &models.Destination{
		Name:      config.Name,
		DestType:  config.Type,
		Config:    config.Config,
		Version:   config.Version,
		ProjectID: projectIDStr,
	}

	// Set user info
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		dest.CreatedBy = user
		dest.UpdatedBy = user
	}

	if err := c.destORM.Create(dest); err != nil {
		return nil, fmt.Errorf("failed to create destination: %s", err)
	}

	return dest, nil
}
