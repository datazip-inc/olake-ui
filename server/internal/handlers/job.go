package handlers

import (
	"context"
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

		query := fmt.Sprintf("WorkflowId between 'sync-%d-%d' and 'sync-%d-%d-~'", projectID, job.ID, projectID, job.ID)
		fmt.Println("Query:", query)
		// List workflows using the direct query
		resp, err := c.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Query:    query,
			PageSize: 1,
		})
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to list workflows: %v", err))
			return
		}

		if len(resp.Executions) > 0 {
			jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
			jobResp.LastRunState = resp.Executions[0].Status.String()
		} else {
			jobResp.LastRunTime = ""
			jobResp.LastRunState = ""
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

	utils.SuccessResponse(&c.Controller, struct {
		Name string `json:"name"`
	}{
		Name: jobName,
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

	utils.SuccessResponse(&c.Controller,
		struct {
			StreamsConfig string `json:"streams_config"`
		}{
			StreamsConfig: job.StreamsConfig,
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

	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.Atoi(projectIDStr)

	if projectIDStr == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
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

	var syncState map[string]interface{}
	if c.tempClient != nil {
		fmt.Println("Using Temporal workflow for sync job")
		syncState, err = c.tempClient.RunSync(
			c.Ctx.Request.Context(),
			job.SourceID.Type,
			job.SourceID.Version,
			job.SourceID.Config,
			job.DestID.Config,
			job.State,
			job.StreamsConfig,
			projectID,
			job.ID,
			job.SourceID.ID,
			job.DestID.ID,
		)
		if err != nil {
			fmt.Printf("Temporal workflow execution failed: %v", err)
		} else {
			fmt.Println("Successfully executed sync job via Temporal")
		}
	}

	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Sync operation failed: %v", err))
		return
	}

	// Update job state with sync result from state.json
	// stateMap := map[string]interface{}{
	// 	"last_run_time":  time.Now().Format(time.RFC3339),
	// 	"last_run_state": "success",
	// 	"sync_state":     syncState,
	// }
	stateJSON, _ := json.Marshal(syncState)
	job.State = string(stateJSON)

	// Update job in database
	if err := c.jobORM.Update(job); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update job state")
		return
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
	var req struct {
		Activate bool `json:"activate"`
	}
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

	utils.SuccessResponse(&c.Controller,
		struct {
			Activate bool `json:"activate"`
		}{
			Activate: job.Active,
		},
	)
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
	projectID, err := strconv.Atoi(projectIDStr)

	if projectIDStr == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}
	// Get job to verify it exists
	job, err := c.jobORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
		return
	}
	var tasks []models.JobTask
	// Construct a query for workflows related to this project and job
	// Using a simpler approach with ExecutionStatus and WorkflowType
	query := fmt.Sprintf("WorkflowId between 'sync-%d-%d' and 'sync-%d-%d-~'", projectID, job.ID, projectID, job.ID)
	fmt.Println("Query:", query)
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

	// Read the log file

	// Get home directory
	homeDir := docker.GetDefaultConfigDir()

	// Check if the main sync directory exists
	mainSyncDir := filepath.Join(homeDir, req.FilePath)
	if _, err := os.Stat(mainSyncDir); os.IsNotExist(err) {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "No sync directory found")
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
	fmt.Println(files)
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
