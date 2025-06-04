package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/docker"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/internal/temporal"
	"github.com/datazip/olake-server/utils"
	"go.temporal.io/api/workflowservice/v1"
)

type JobHandler struct {
	web.Controller
	jobORM     *database.JobORM
	sourceORM  *database.SourceORM
	destORM    *database.DestinationORM
	tempClient *temporal.Client
}

// Prepare initializes the ORM instances
func (c *JobHandler) Prepare() {
	c.jobORM = database.NewJobORM()
	c.sourceORM = database.NewSourceORM()
	c.destORM = database.NewDestinationORM()
	tempAddress := web.AppConfig.DefaultString("TEMPORAL_ADDRESS", "localhost:7233")
	tempClient, err := temporal.NewClient(tempAddress)
	if err != nil {
		// Log the error but continue - we'll fall back to direct Docker execution if Temporal fails
		logs.Error("Failed to create Temporal client: %v", err)
	} else {
		c.tempClient = tempClient
	}
}

// @router /project/:projectid/jobs [get]
func (c *JobHandler) GetAllJobs() {
	projectIDStr := c.Ctx.Input.Param(":projectid")
	// Get jobs with optional filtering
	jobs, err := c.jobORM.GetAllByProjectID(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs by project ID")
		return
	}

	// Transform to response format
	jobResponses := make([]models.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		jobResp := models.JobResponse{
			ID:            job.ID,
			Name:          job.Name,
			StreamsConfig: job.StreamsConfig,
			Frequency:     job.Frequency,
			CreatedAt:     job.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     job.UpdatedAt.Format(time.RFC3339),
			Activate:      job.Active,
		}

		// Set source and destination details
		if job.SourceID != nil {
			jobResp.Source = models.JobSourceConfig{
				Name:    job.SourceID.Name,
				Type:    job.SourceID.Type,
				Config:  job.SourceID.Config,
				Version: job.SourceID.Version,
			}
		}

		if job.DestID != nil {
			jobResp.Destination = models.JobDestinationConfig{
				Name:    job.DestID.Name,
				Type:    job.DestID.DestType,
				Config:  job.DestID.Config,
				Version: job.DestID.Version,
			}
		}

		// Set user details
		if job.CreatedBy != nil {
			jobResp.CreatedBy = job.CreatedBy.Username
		}
		if job.UpdatedBy != nil {
			jobResp.UpdatedBy = job.UpdatedBy.Username
		}

		// Get workflow information if Temporal client is available
		if c.tempClient != nil {
			query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectIDStr, job.ID, projectIDStr, job.ID)
			if resp, err := c.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
				Query:    query,
				PageSize: 1,
			}); err != nil {
				logs.Error("Failed to list workflows: %v", err)
			} else if len(resp.Executions) > 0 {
				jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
				jobResp.LastRunState = resp.Executions[0].Status.String()
			}
		}

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
		_, err = c.tempClient.CreateSync(
			c.Ctx.Request.Context(),
			job.Frequency,
			job.ProjectID,
			job.ID,
			false,
		)
		if err != nil {
			fmt.Printf("Temporal workflow execution failed: %v", err)
		} else {
			fmt.Println("Successfully executed sync job via Temporal")
		}
	}

	utils.SuccessResponse(&c.Controller, req)
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
		fmt.Println("Using Temporal workflow for sync job")
		_, err = c.tempClient.CreateSync(
			c.Ctx.Request.Context(),
			existingJob.Frequency,
			existingJob.ProjectID,
			existingJob.ID,
			false,
		)
		if err != nil {
			fmt.Printf("Temporal workflow execution failed: %v", err)
		} else {
			fmt.Println("Successfully executed sync job via Temporal")
		}
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

	utils.SuccessResponse(&c.Controller, models.DeleteDestinationResponse{
		Name: jobName,
	})
}

// no need any more
// @router /project/:projectid/jobs/:id/streams [get]
// func (c *JobHandler) GetJobStreams() {
// 	idStr := c.Ctx.Input.Param(":id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
// 		return
// 	}

// 	// Get job
// 	job, err := c.jobORM.GetByID(id)
// 	if err != nil {
// 		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
// 		return
// 	}

// 	utils.SuccessResponse(&c.Controller,
// 		struct {
// 			StreamsConfig string `json:"streams_config"`
// 		}{
// 			StreamsConfig: job.StreamsConfig,
// 		},
// 	)
// }

// @router /project/:projectid/jobs/:id/sync [post]
func (c *JobHandler) SyncJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid job ID")
		return
	}

	projectIDStr := c.Ctx.Input.Param(":projectid")
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

	if c.tempClient != nil {
		fmt.Println("Using Temporal workflow for sync job")
		resp, err := c.tempClient.CreateSync(
			c.Ctx.Request.Context(),
			job.Frequency,
			projectIDStr,
			job.ID,
			true,
		)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("temporal excuetion failed: %v", err))
		} else {
			utils.SuccessResponse(&c.Controller, resp)
		}
	}
	utils.SuccessResponse(&c.Controller, nil)
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

	// Get existing job
	job, err := c.jobORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
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
	projectIDStr := c.Ctx.Input.Param(":projectid")

	// Get job to verify it exists
	job, err := c.jobORM.GetByID(id)
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
		var runTime time.Duration
		var endTime time.Time
		startTime := execution.StartTime.AsTime()

		if execution.CloseTime != nil {
			endTime = execution.CloseTime.AsTime()
			runTime = endTime.Sub(startTime)
		}
		tasks = append(tasks, models.JobTask{
			Runtime:   runTime.String(),
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

	// Verify job exists
	_, err = c.jobORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(req.FilePath)))
	// Read the log file

	// Get home directory
	homeDir := docker.GetDefaultConfigDir()
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	if _, err := os.Stat(mainSyncDir); os.IsNotExist(err) {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, fmt.Sprintf("No sync directory found: %s", mainSyncDir))
		return
	}

	// Look for log files in the logs directory
	logsDir := filepath.Join(mainSyncDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Logs directory not found")
		return
	}

	// Since there is only one sync folder in logs, we can get it directly
	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "No sync log directory found")
		return
	}

	// Use the first directory we find (since there's only one)
	syncDir := filepath.Join(logsDir, files[0].Name())

	// Define the log file path
	logPath := filepath.Join(syncDir, "olake.log")

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to read log file : %s", logPath))
		return
	}

	// Parse log entries
	var logs []map[string]interface{}
	lines := strings.Split(string(logContent), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var logEntry struct {
			Level   string    `json:"level"`
			Time    time.Time `json:"time"`
			Message string    `json:"message"`
		}

		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		logs = append(logs, map[string]interface{}{
			"level":   logEntry.Level,
			"time":    logEntry.Time.Format(time.RFC3339),
			"message": logEntry.Message,
		})
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
