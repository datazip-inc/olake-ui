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
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
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
	logs.Info("Creating job service")
	tempClient, err := temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
		//Q: return nil or error?
	}
	return &JobService{
		jobORM:     database.NewJobORM(),
		sourceORM:  database.NewSourceORM(),
		destORM:    database.NewDestinationORM(),
		tempClient: tempClient,
	}, nil
}

func (s *JobService) GetAllJobsByProject(projectID string) ([]models.JobResponse, error) {
	logs.Info("Retrieving jobs by project ID: %s", projectID)
	jobs, err := s.jobORM.GetAllByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("%s jobs by project ID: %s", constants.ErrFailedToRetrieve, err)
	}

	jobResponses := make([]models.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		jobResp := s.buildJobResponse(job, projectID)
		jobResponses = append(jobResponses, jobResp)
	}

	return jobResponses, nil
}

func (s *JobService) CreateJob(ctx context.Context, req *models.CreateJobRequest, projectID string, userID *int) error {
	logs.Info("Creating job: %s", req.Name)
	source, err := s.getOrCreateSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source: %s", err)
	}

	dest, err := s.getOrCreateDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination: %s", err)
	}

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

	if userID != nil {
		user := &models.User{ID: *userID}
		job.CreatedBy = user
		job.UpdatedBy = user
	}

	if err := s.jobORM.Create(job); err != nil {
		return fmt.Errorf("%s job: %s", constants.ErrFailedToCreate, err)
	}

	if s.tempClient != nil {
		logs.Info("Creating Temporal workflow for sync job")
		_, err = s.tempClient.ManageSync(ctx, job.ProjectID, job.ID, job.Frequency, temporal.ActionCreate)
		if err != nil {
			logs.Error("%s: %v", constants.ErrWorkflowExecutionFailed, err)
		} else {
			logs.Info("Successfully created sync job via Temporal")
		}
	}
	telemetry.TrackJobCreation(ctx, &models.Job{Name: req.Name})
	return nil
}

func (s *JobService) UpdateJob(ctx context.Context, req *models.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	logs.Info("Updating job: %s", req.Name)
	existingJob, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return fmt.Errorf("job not found: %s", err)
	}

	source, err := s.getOrCreateSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source: %s", err)
	}

	dest, err := s.getOrCreateDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination: %s", err)
	}

	existingJob.Name = req.Name
	existingJob.SourceID = source
	existingJob.DestID = dest
	existingJob.Active = req.Activate
	existingJob.Frequency = req.Frequency
	existingJob.StreamsConfig = req.StreamsConfig
	existingJob.UpdatedAt = time.Now()
	existingJob.ProjectID = projectID

	if userID != nil {
		user := &models.User{ID: *userID}
		existingJob.UpdatedBy = user
	}

	if err := s.jobORM.Update(existingJob); err != nil {
		return fmt.Errorf("failed to update job: %s", err)
	}

	if s.tempClient != nil {
		logs.Info("Updating Temporal workflow for sync job")
		_, err = s.tempClient.ManageSync(ctx, existingJob.ProjectID, existingJob.ID, existingJob.Frequency, temporal.ActionUpdate)
		if err != nil {
			return fmt.Errorf("temporal workflow execution failed: %s", err)
		}
	}
	telemetry.TrackJobEntity(ctx)
	return nil
}

func (s *JobService) DeleteJob(ctx context.Context, jobID int) (string, error) {
	logs.Info("Deleting job with id: %d", jobID)
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return "", fmt.Errorf("job not found: %s", err)
	}

	jobName := job.Name

	if s.tempClient != nil {
		logs.Info("Deleting Temporal workflow")
		_, err := s.tempClient.ManageSync(ctx, job.ProjectID, job.ID, job.Frequency, temporal.ActionDelete)
		if err != nil {
			logs.Error("Temporal deletion failed: %v", err)
		}
	}

	if err := s.jobORM.Delete(jobID); err != nil {
		return "", fmt.Errorf("failed to delete job: %s", err)
	}
	telemetry.TrackJobEntity(ctx)
	return jobName, nil
}

func (s *JobService) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	logs.Info("Syncing job with id: %d", jobID)
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", err)
	}

	if job.SourceID == nil || job.DestID == nil {
		return nil, fmt.Errorf("job must have both source and destination configured")
	}

	if s.tempClient != nil {
		resp, err := s.tempClient.ManageSync(ctx, job.ProjectID, job.ID, job.Frequency, temporal.ActionTrigger)
		if err != nil {
			return nil, fmt.Errorf("temporal execution failed: %s", err)
		}
		return resp, nil
	}

	return nil, fmt.Errorf("temporal client is not available")
}

func (s *JobService) ActivateJob(jobID int, activate bool, userID *int) error {
	logs.Info("Activating job with id: %d", jobID)
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return fmt.Errorf("job not found: %s", err)
	}

	job.Active = activate
	job.UpdatedAt = time.Now()

	if userID != nil {
		user := &models.User{ID: *userID}
		job.UpdatedBy = user
	}

	if err := s.jobORM.Update(job); err != nil {
		return fmt.Errorf("failed to update job activation status: %s", err)
	}

	return nil
}

func (s *JobService) GetJobTasks(ctx context.Context, projectID string, jobID int) ([]models.JobTask, error) {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", err)
	}

	if s.tempClient == nil {
		return []models.JobTask{}, nil
	}

	var tasks []models.JobTask
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)

	resp, err := s.tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %s", err)
	}

	for _, execution := range resp.Executions {
		var runTime time.Duration
		startTime := execution.StartTime.AsTime()
		if execution.CloseTime != nil {
			runTime = execution.CloseTime.AsTime().Sub(startTime)
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

func (s *JobService) GetTaskLogs(ctx context.Context, jobID int, filePath string) ([]map[string]interface{}, error) {
	logs.Info("Getting task logs for job with id: %d", jobID)
	_, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", err)
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))
	mainSyncDir := filepath.Join(docker.DefaultConfigDir, syncFolderName)
	if _, err := os.Stat(mainSyncDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no sync directory found: %s", mainSyncDir)
	}

	logsDir := filepath.Join(mainSyncDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("logs directory not found")
	}

	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no sync log directory found")
	}

	syncDir := filepath.Join(logsDir, files[0].Name())
	logPath := filepath.Join(syncDir, "olake.log")

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %s", logPath)
	}

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

	if job.CreatedBy != nil {
		jobResp.CreatedBy = job.CreatedBy.Username
	}
	if job.UpdatedBy != nil {
		jobResp.UpdatedBy = job.UpdatedBy.Username
	}

	if s.tempClient != nil {
		query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)
		if resp, err := s.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Query:    query,
			PageSize: 1,
		}); err == nil && len(resp.Executions) > 0 {
			jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
			jobResp.LastRunState = resp.Executions[0].Status.String()
		}
	}

	return jobResp
}

func (s *JobService) getOrCreateSource(config models.JobSourceConfig, projectID string, userID *int) (*models.Source, error) {
	sources, err := s.sourceORM.GetByNameAndType(config.Name, config.Type, projectID)
	if err == nil && len(sources) > 0 {
		source := sources[0]
		source.Config = config.Config
		source.Version = config.Version
		if userID != nil {
			source.UpdatedBy = &models.User{ID: *userID}
		}
		if err := s.sourceORM.Update(source); err != nil {
			return nil, err
		}
		return source, nil
	}
	source := &models.Source{
		Name:      config.Name,
		Type:      config.Type,
		Config:    config.Config,
		Version:   config.Version,
		ProjectID: projectID,
	}
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
	destinations, err := s.destORM.GetByNameAndType(config.Name, config.Type, projectID)
	if err == nil && len(destinations) > 0 {
		dest := destinations[0]
		dest.Config = config.Config
		dest.Version = config.Version
		if userID != nil {
			dest.UpdatedBy = &models.User{ID: *userID}
		}
		if err := s.destORM.Update(dest); err != nil {
			return nil, err
		}
		return dest, nil
	}
	dest := &models.Destination{
		Name:      config.Name,
		DestType:  config.Type,
		Config:    config.Config,
		Version:   config.Version,
		ProjectID: projectID,
	}
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
