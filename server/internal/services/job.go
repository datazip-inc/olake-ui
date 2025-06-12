package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/temporal"
	"go.temporal.io/api/workflowservice/v1"
)

type JobService struct {
	jobORM     *database.JobORM
	sourceORM  *database.SourceORM
	destORM    *database.DestinationORM
	tempClient *temporal.Client
}

func NewJobService() (*JobService, error) {
	tempClient, err := temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
		// Don't return error, allow service to work without Temporal
	}

	return &JobService{
		jobORM:     database.NewJobORM(),
		sourceORM:  database.NewSourceORM(),
		destORM:    database.NewDestinationORM(),
		tempClient: tempClient,
	}, nil
}

func (s *JobService) GetAllJobsByProject(projectID string) ([]models.JobResponse, error) {
	jobs, err := s.jobORM.GetAllByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs by project ID: %w", err)
	}

	jobResponses := make([]models.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		jobResp := s.buildJobResponse(job, projectID)
		jobResponses = append(jobResponses, jobResp)
	}

	return jobResponses, nil
}

func (s *JobService) CreateJob(ctx context.Context, req *models.CreateJobRequest, projectID string, userID *int) error {
	// Find or create source
	source, err := s.getOrCreateSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source: %w", err)
	}

	// Find or create destination
	dest, err := s.getOrCreateDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination: %w", err)
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
		ProjectID:     projectID,
	}

	// Set user information
	if userID != nil {
		user := &models.User{ID: *userID}
		job.CreatedBy = user
		job.UpdatedBy = user
	}

	// Create job in database
	if err := s.jobORM.Create(job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Create Temporal workflow if client is available
	if s.tempClient != nil {
		logs.Info("Creating Temporal workflow for sync job")
		_, err = s.tempClient.CreateSync(ctx, job.Frequency, job.ProjectID, job.ID, false)
		if err != nil {
			logs.Error("Temporal workflow execution failed: %v", err)
		} else {
			logs.Info("Successfully created sync job via Temporal")
		}
	}

	return nil
}

func (s *JobService) UpdateJob(ctx context.Context, req *models.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	// Get existing job
	existingJob, err := s.jobORM.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// Find or create source
	source, err := s.getOrCreateSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source: %w", err)
	}

	// Find or create destination
	dest, err := s.getOrCreateDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination: %w", err)
	}

	// Update fields
	existingJob.Name = req.Name
	existingJob.SourceID = source
	existingJob.DestID = dest
	existingJob.Active = req.Activate
	existingJob.Frequency = req.Frequency
	existingJob.StreamsConfig = req.StreamsConfig
	existingJob.UpdatedAt = time.Now()
	existingJob.ProjectID = projectID

	// Update user information
	if userID != nil {
		user := &models.User{ID: *userID}
		existingJob.UpdatedBy = user
	}

	// Update job in database
	if err := s.jobORM.Update(existingJob); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Update Temporal workflow if client is available
	if s.tempClient != nil {
		logs.Info("Updating Temporal workflow for sync job")
		_, err = s.tempClient.CreateSync(ctx, existingJob.Frequency, existingJob.ProjectID, existingJob.ID, false)
		if err != nil {
			return fmt.Errorf("temporal workflow execution failed: %w", err)
		}
	}

	return nil
}

func (s *JobService) DeleteJob(jobID int) (string, error) {
	// Get job name for response
	job, err := s.jobORM.GetByID(jobID)
	if err != nil {
		return "", fmt.Errorf("job not found: %w", err)
	}

	jobName := job.Name

	// Delete job
	if err := s.jobORM.Delete(jobID); err != nil {
		return "", fmt.Errorf("failed to delete job: %w", err)
	}

	return jobName, nil
}

func (s *JobService) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	// Check if job exists
	job, err := s.jobORM.GetByID(jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	// Validate source and destination exist
	if job.SourceID == nil || job.DestID == nil {
		return nil, fmt.Errorf("job must have both source and destination configured")
	}

	if s.tempClient != nil {
		logs.Info("Using Temporal workflow for sync job")
		resp, err := s.tempClient.CreateSync(ctx, job.Frequency, projectID, job.ID, true)
		if err != nil {
			return nil, fmt.Errorf("temporal execution failed: %w", err)
		}
		return resp, nil
	}

	return nil, fmt.Errorf("temporal client is not available")
}

func (s *JobService) ActivateJob(jobID int, activate bool, userID *int) error {
	// Get existing job
	job, err := s.jobORM.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// Update activation status
	job.Active = activate
	job.UpdatedAt = time.Now()

	// Update user information
	if userID != nil {
		user := &models.User{ID: *userID}
		job.UpdatedBy = user
	}

	// Update job in database
	if err := s.jobORM.Update(job); err != nil {
		return fmt.Errorf("failed to update job activation status: %w", err)
	}

	return nil
}

func (s *JobService) GetJobTasks(projectID string, jobID int) ([]models.JobTask, error) {
	// Get job to verify it exists
	job, err := s.jobORM.GetByID(jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	if s.tempClient == nil {
		return []models.JobTask{}, nil
	}

	var tasks []models.JobTask
	// Construct a query for workflows related to this project and job
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)

	// List workflows using the direct query
	resp, err := s.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	for _, execution := range resp.Executions {
		var runTime time.Duration
		startTime := execution.StartTime.AsTime()

		if execution.CloseTime != nil {
			endTime := execution.CloseTime.AsTime()
			runTime = endTime.Sub(startTime)
		}

		tasks = append(tasks, models.JobTask{
			Runtime:   runTime.String(),
			StartTime: startTime.UTC().Format(time.RFC3339),
			Status:    execution.Status.String(),
			FilePath:  execution.Execution.WorkflowId,
		})
	}

	return tasks, nil
}

func (s *JobService) GetTaskLogs(jobID int, filePath string) ([]map[string]interface{}, error) {
	// Verify job exists
	_, err := s.jobORM.GetByID(jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))

	// Get home directory
	mainSyncDir := filepath.Join(docker.DefaultConfigDir, syncFolderName)
	if _, err := os.Stat(mainSyncDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no sync directory found: %s", mainSyncDir)
	}

	// Look for log files in the logs directory
	logsDir := filepath.Join(mainSyncDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("logs directory not found")
	}

	// Since there is only one sync folder in logs, we can get it directly
	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no sync log directory found")
	}

	// Use the first directory we find (since there's only one)
	syncDir := filepath.Join(logsDir, files[0].Name())

	// Define the log file path
	logPath := filepath.Join(syncDir, "olake.log")

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %s", logPath)
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
			"time":    logEntry.Time.UTC().Format(time.RFC3339),
			"message": logEntry.Message,
		})
	}

	return logs, nil
}

// Private helper methods

func (s *JobService) buildJobResponse(job *models.Job, projectID string) models.JobResponse {
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
	if s.tempClient != nil {
		query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)
		if resp, err := s.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Query:    query,
			PageSize: 1,
		}); err != nil {
			logs.Error("Failed to list workflows: %v", err)
		} else if len(resp.Executions) > 0 {
			jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
			jobResp.LastRunState = resp.Executions[0].Status.String()
		}
	}

	return jobResp
}

func (s *JobService) getOrCreateSource(config models.JobSourceConfig, projectID string, userID *int) (*models.Source, error) {
	// Try to find an existing source matching the criteria
	sources, err := s.sourceORM.GetByNameAndType(config.Name, config.Type, projectID)
	if err == nil && len(sources) > 0 {
		// Update the existing source if found
		source := sources[0]
		source.Config = config.Config
		source.Version = config.Version

		// Get user info for update
		if userID != nil {
			user := &models.User{ID: *userID}
			source.UpdatedBy = user
		}

		if err := s.sourceORM.Update(source); err != nil {
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
		ProjectID: projectID,
	}

	// Set user info
	if userID != nil {
		user := &models.User{ID: *userID}
		source.CreatedBy = user
		source.UpdatedBy = user
	}

	if err := s.sourceORM.Create(source); err != nil {
		return nil, err
	}

	return source, nil
}

func (s *JobService) getOrCreateDestination(config models.JobDestinationConfig, projectID string, userID *int) (*models.Destination, error) {
	// Try to find an existing destination matching the criteria
	destinations, err := s.destORM.GetByNameAndType(config.Name, config.Type, projectID)
	if err == nil && len(destinations) > 0 {
		// Update the existing destination if found
		dest := destinations[0]
		dest.Config = config.Config
		dest.Version = config.Version

		// Get user info for update
		if userID != nil {
			user := &models.User{ID: *userID}
			dest.UpdatedBy = user
		}

		if err := s.destORM.Update(dest); err != nil {
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
		ProjectID: projectID,
	}

	// Set user info
	if userID != nil {
		user := &models.User{ID: *userID}
		dest.CreatedBy = user
		dest.UpdatedBy = user
	}

	if err := s.destORM.Create(dest); err != nil {
		return nil, err
	}

	return dest, nil
}
