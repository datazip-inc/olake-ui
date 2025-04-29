package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/docker"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/utils"
)

type JobHandler struct {
	web.Controller
	jobORM    *database.JobORM
	sourceORM *database.SourceORM
	destORM   *database.DestinationORM
}

// Prepare initializes the ORM instances
func (c *JobHandler) Prepare() {
	c.jobORM = database.NewJobORM()
	c.sourceORM = database.NewSourceORM()
	c.destORM = database.NewDestinationORM()
}

// @router /project/:projectid/jobs [get]
func (c *JobHandler) GetAllJobs() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Get optional query parameters for filtering
	sourceID := c.GetString("source_id")
	destID := c.GetString("dest_id")

	var jobs []*models.Job
	var getErr error

	// Apply filters if provided
	if sourceID != "" {
		sourceIDInt, err := strconv.Atoi(sourceID)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
			return
		}
		jobs, getErr = c.jobORM.GetBySourceID(sourceIDInt)
	} else if destID != "" {
		destIDInt, err := strconv.Atoi(destID)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
			return
		}
		jobs, getErr = c.jobORM.GetByDestinationID(destIDInt)
	} else {
		// Get jobs for the project
		jobs, getErr = c.jobORM.GetAllByProjectID(projectID)
	}

	if getErr != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs")
		return
	}

	// Transform to response format
	jobResponses := make([]models.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		// Get source and destination details
		source := job.SourceID
		dest := job.DestID

		// Create response object
		jobResp := models.JobResponse{
			ID:            job.ID,
			Name:          job.Name,
			StreamsConfig: job.StreamsConfig,
			Frequency:     job.Frequency,
			CreatedAt:     job.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     job.UpdatedAt.Format(time.RFC3339),
			Activate:      job.Active,
		}

		// Set source details if available
		if source != nil {
			jobResp.Source = models.JobSourceConfig{
				Name:    source.Name,
				Type:    source.Type,
				Config:  source.Config,
				Version: source.Version,
			}
		}

		// Set destination details if available
		if dest != nil {
			jobResp.Destination = models.JobDestinationConfig{
				Name:    dest.Name,
				Type:    dest.DestType,
				Config:  dest.Config,
				Version: dest.Version,
			}
		}

		// Set user details if available
		if job.CreatedBy != nil {
			jobResp.CreatedBy = job.CreatedBy.Username
		}

		if job.UpdatedBy != nil {
			jobResp.UpdatedBy = job.UpdatedBy.Username
		}

		// Parse state for last run info if available
		// if job.State != "" {
		// 	var state map[string]interface{}
		// 	if err := json.Unmarshal([]byte(job.State), &state); err == nil {
		// 		if lastRunTime, ok := state["last_run_time"].(string); ok {
		// 			jobResp.LastRunTime = lastRunTime
		// 		}

		// 		if lastRunState, ok := state["last_run_state"].(string); ok {
		// 			jobResp.LastRunState = lastRunState
		// 		}
		// 	}
		// }
		jobResp.LastRunTime = "2025-04-27T15:30:00Z"
		jobResp.LastRunState = "success"

		jobResponses = append(jobResponses, jobResp)
	}

	utils.SuccessResponse(&c.Controller, jobResponses)
}

// @router /project/:projectid/jobs [post]
func (c *JobHandler) CreateJob() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")

	// Parse request body
	var req models.CreateJobRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
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

	// Create job model
	job := &models.Job{
		Name:          req.Name,
		SourceID:      source,
		DestID:        dest,
		Active:        true,
		Frequency:     req.Frequency,
		StreamsConfig: req.StreamsConfig,
		State:         "{}",
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

	utils.SuccessResponse(&c.Controller, models.CreateJobResponse{
		Success: true,
		Message: "Job created successfully",
		Data:    req,
	})
}

// @router /project/:projectid/jobs/:id [put]
func (c *JobHandler) UpdateJob() {
	// Get project ID and job ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
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

	// Get existing job
	existingJob, err := c.jobORM.GetByID(id)
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

	utils.SuccessResponse(&c.Controller, models.UpdateJobResponse{
		Success: true,
		Message: "Job updated successfully",
		Data:    req,
	})
}

// @router /project/:projectid/jobs/:id [delete]
func (c *JobHandler) DeleteJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Get job name for response
	job, err := c.jobORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}

	jobName := job.Name

	// Delete job
	if err := c.jobORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete job")
		return
	}

	utils.SuccessResponse(&c.Controller, models.DeleteJobResponse{
		Success: true,
		Message: "Job deleted successfully",
		Data: struct {
			Name string `json:"name"`
		}{
			Name: jobName,
		},
	})
}

// @router /project/:projectid/jobs/:id/streams [get]
func (c *JobHandler) GetJobStreams() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Get job
	job, err := c.jobORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}

	utils.SuccessResponse(&c.Controller, models.GetJobStreamsResponse{
		Success: true,
		Message: "Job streams retrieved successfully",
		Data: struct {
			StreamsConfig string `json:"streams_config"`
		}{
			StreamsConfig: job.StreamsConfig,
		},
	},
	)
}

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Check if job exists
	job, err := c.jobORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}

	// Validate source and destination exist
	if job.SourceID == nil || job.DestID == nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Job must have both source and destination configured")
		return
	}

	// Create Docker runner
	configDir := docker.GetDefaultConfigDir()
	runner := docker.NewRunner(configDir)

	// Run sync operation - the RunSync method will generate the catalog automatically if needed
	syncState, err := runner.RunSync(
		job.SourceID.Type,
		job.SourceID.Version,
		job.SourceID.Config,
		job.DestID.Config,
		job.SourceID.ID,
		job.DestID.ID,
	)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Sync operation failed: %v", err))
		return
	}

	// Update job state with sync result from state.json
	stateMap := map[string]interface{}{
		"last_run_time":  time.Now().Format(time.RFC3339),
		"last_run_state": "success",
		"sync_state":     syncState,
	}
	stateJSON, _ := json.Marshal(stateMap)
	job.State = string(stateJSON)

	// Update job in database
	if err := c.jobORM.Update(job); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job state")
		return
	}

	utils.SuccessResponse(&c.Controller, struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Success: true,
		Message: "Job sync completed successfully",
		Data:    syncState,
	})
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
			return nil, err
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
		return nil, err
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
			return nil, err
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
		return nil, err
	}

	return dest, nil
}
