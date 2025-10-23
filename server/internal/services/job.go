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
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/internal/temporal"
	"github.com/datazip/olake-ui/server/utils"
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

func (s *JobService) GetAllJobs(projectID string) ([]dto.JobResponse, error) {
	logs.Info("Retrieving jobs by project ID: %s", projectID)
	jobs, err := s.jobORM.GetAllJobsByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("%s jobs by project ID: %s", constants.ErrFailedToRetrieve, err)
	}

	jobResponses := make([]dto.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		jobResp := s.buildJobResponse(job, projectID)
		jobResponses = append(jobResponses, jobResp)
	}

	return jobResponses, nil
}

func (s *JobService) CreateJob(ctx context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error {
	isUnique, err := s.jobORM.IsJobNameUnique(projectID, req.Name)
	if err != nil {
		return fmt.Errorf("failed to check job uniqness")
	}

	if !isUnique {
		return fmt.Errorf("job name already exists")
	}

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
		logs.Info("Creating Temporal workflow for sync job for job id %d", job.ID)
		_, err = s.tempClient.ManageSync(ctx, job, temporal.ActionCreate)
		if err != nil {
			return fmt.Errorf("%s: %s", constants.ErrWorkflowExecutionFailed, err)
		}
		logs.Info("Successfully created sync job via Temporal for job id %d", job.ID)
	}
	telemetry.TrackJobCreation(ctx, job)
	return nil
}

func (s *JobService) UpdateJob(ctx context.Context, req *dto.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	logs.Info("Updating job: %s", req.Name)
	existingJob, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return fmt.Errorf("job not found: %s", err)
	}

	// block when clear-destination is running
	clearRunning, err := isClearRunning(ctx, s.tempClient, projectID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get clear destination status: %s", err)
	}
	if clearRunning {
		return constants.ErrInProgress
	}

	// cancel sync before updating the job
	syncRunning, err := isSyncRunning(ctx, s.tempClient, projectID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}
	if syncRunning {
		logs.Info("sync is running for job %d, initiating cancel sync workflow", jobID)
		if err := cancelJobWorkflow(ctx, s.tempClient, projectID, jobID); err != nil {
			return err
		}
		logs.Info("successfully cancelled sync for job %d", jobID)
	}

	// start clear destination
	var diffCatalog map[string]interface{}
	if err := json.Unmarshal([]byte(req.DifferenceStreams), &diffCatalog); err != nil {
		return fmt.Errorf("invalid difference_streams JSON: %s", err)
	}

	if len(diffCatalog) != 0 {
		logs.Info("Stream difference detected for job %d, running clear destination workflow", existingJob.ID)
		if _, err := s.ClearDestination(ctx, projectID, jobID, req.DifferenceStreams); err != nil {
			return fmt.Errorf("failed to run clear destination workflow: %s", err)
		}
		logs.Info("Successfully triggered clear destination workflow for job %d", existingJob.ID)
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
	existingJob.ProjectID = projectID

	if userID != nil {
		user := &models.User{ID: *userID}
		existingJob.UpdatedBy = user
	}

	if err := s.jobORM.Update(existingJob); err != nil {
		return fmt.Errorf("failed to update job: %s", err)
	}

	if s.tempClient != nil {
		logs.Info("Updating Temporal workflow for sync job id %d", existingJob.ID)
		_, err = s.tempClient.ManageSync(ctx, existingJob, temporal.ActionUpdate)
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
		return "", fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	jobName := job.Name

	if s.tempClient != nil {
		logs.Info("Deleting Temporal workflow for sync job id %d", job.ID)
		_, err := s.tempClient.ManageSync(ctx, job, temporal.ActionDelete)
		if err != nil {
			logs.Error("Temporal deletion failed: %v", err)
		}
	}

	if err := s.jobORM.Delete(jobID); err != nil {
		return "", fmt.Errorf("failed to delete job id %d: %s", jobID, err)
	}
	telemetry.TrackJobEntity(ctx)
	return jobName, nil
}

func (s *JobService) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	logs.Info("Syncing job with id: %d", jobID)
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	if job.SourceID == nil || job.DestID == nil {
		return nil, fmt.Errorf("job must have both source and destination configured")
	}

	running, err := isClearRunning(ctx, s.tempClient, projectID, jobID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get clear destination status: %s", err)
	}
	if running {
		return nil, constants.ErrInProgress
	}

	if s.tempClient != nil {
		resp, err := s.tempClient.ManageSync(ctx, job, temporal.ActionTrigger)
		if err != nil {
			return nil, fmt.Errorf("temporal execution failed: %s", err)
		}
		return resp, nil
	}

	return nil, fmt.Errorf("temporal client is not available")
}

func (s *JobService) CancelJobRun(ctx context.Context, projectID string, jobID int) (map[string]any, error) {
	if _, err := s.jobORM.GetByID(jobID, true); err != nil {
		return nil, fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	if err := cancelJobWorkflow(ctx, s.tempClient, projectID, jobID); err != nil {
		return nil, fmt.Errorf("job workflow cancel failed id %d: %s", jobID, err)
	}
	return map[string]any{
		"message": "job workflow cancel requested successfully",
	}, nil
}

func (s *JobService) ActivateJob(ctx context.Context, jobID int, activate bool, userID *int) error {
	logs.Info("Activating job with id: %d", jobID)
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	job.Active = activate

	if userID != nil {
		user := &models.User{ID: *userID}
		job.UpdatedBy = user
	}

	if err := s.jobORM.Update(job); err != nil {
		return fmt.Errorf("failed to update job activation status: %s", err)
	}

	if _, err := s.tempClient.ManageSync(ctx, job, utils.Ternary(activate, temporal.ActionUnpause, temporal.ActionPause).(temporal.SyncAction)); err != nil {
		return fmt.Errorf("failed to update job activation status: %s", err)
	}

	return nil
}

func (s *JobService) StreamsDifference(ctx context.Context, projectID string, jobID int, req dto.StreamDifferenceRequest) (map[string]interface{}, error) {
	existingJob, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	diffCatalog, err := s.tempClient.GetStreamDiff(ctx, existingJob, existingJob.StreamsConfig, req.UpdatedStreamsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream difference: %s", err)
	}

	return diffCatalog, nil
}

func (s *JobService) IsJobNameUnique(ctx context.Context, projectID string, req dto.CheckUniqueJobNameRequest) (bool, error) {
	logs.Info("Checking if job name is unique: %s", req.JobName)
	return s.jobORM.IsJobNameUnique(projectID, req.JobName)
}

func (s *JobService) GetJobTasks(ctx context.Context, projectID string, jobID int) ([]dto.JobTask, error) {
	if _, err := s.jobORM.GetByID(jobID, true); err != nil {
		return nil, fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	if s.tempClient == nil {
		return []dto.JobTask{}, nil
	}

	var tasks []dto.JobTask
	syncQuery := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, jobID, projectID, jobID)
	clearQuery := fmt.Sprintf("WorkflowId between 'clear-destination-%s-%d' and 'clear-destination-%s-%d-~'", projectID, jobID, projectID, jobID)

	query := fmt.Sprintf("(%s) OR (%s)", syncQuery, clearQuery)

	resp, err := s.tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %s", err)
	}

	for _, execution := range resp.Executions {
		startTime := execution.StartTime.AsTime().UTC()
		var runTime string
		if execution.CloseTime != nil {
			runTime = execution.CloseTime.AsTime().UTC().Sub(startTime).Round(time.Second).String()
		} else {
			runTime = time.Since(startTime).Round(time.Second).String()
		}

		workflowID := execution.Execution.WorkflowId
		jobType := "sync"
		if strings.HasPrefix(workflowID, "clear-destination-") {
			jobType = "clear"
		}

		tasks = append(tasks, dto.JobTask{
			Runtime:   runTime,
			StartTime: startTime.Format(time.RFC3339),
			Status:    execution.Status.String(),
			FilePath:  workflowID,
			JobType:   jobType,
		})
	}

	return tasks, nil
}

func (s *JobService) GetTaskLogs(ctx context.Context, jobID int, filePath string) ([]map[string]interface{}, error) {
	logs.Info("Getting task logs for job with id: %d", jobID)
	if _, err := s.jobORM.GetByID(jobID, true); err != nil {
		return nil, fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))
	mainSyncDir := filepath.Join(constants.DefaultConfigDir, syncFolderName)

	if _, err := os.Stat(mainSyncDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no sync directory found: %s", mainSyncDir)
	}

	logsDir := filepath.Join(mainSyncDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("logs directory not found for job id %d", jobID)
	}

	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no sync log directory found for job id %d", jobID)
	}

	syncDir := filepath.Join(logsDir, files[0].Name())
	logPath := filepath.Join(syncDir, "olake.log")

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file for job id %d: %s", jobID, logPath)
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

func (s *JobService) buildJobResponse(job *models.Job, projectID string) dto.JobResponse {
	jobResp := dto.JobResponse{
		ID:            job.ID,
		Name:          job.Name,
		StreamsConfig: job.StreamsConfig,
		Frequency:     job.Frequency,
		CreatedAt:     job.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     job.UpdatedAt.Format(time.RFC3339),
		Activate:      job.Active,
	}

	if job.SourceID != nil {
		jobResp.Source = dto.JobSourceConfig{
			Name:    job.SourceID.Name,
			Type:    job.SourceID.Type,
			Config:  job.SourceID.Config,
			Version: job.SourceID.Version,
		}
	}

	if job.DestID != nil {
		jobResp.Destination = dto.JobDestinationConfig{
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
		syncQuery := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)
		clearQuery := fmt.Sprintf("WorkflowId between 'clear-destination-%s-%d' and 'clear-destination-%s-%d-~'", projectID, job.ID, projectID, job.ID)
		query := fmt.Sprintf("(%s) OR (%s)", syncQuery, clearQuery)

		resp, err := s.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Query:    query,
			PageSize: 1,
		})
		if err != nil {
			logs.Error("Failed to list workflows: %s", err)
		}

		if len(resp.Executions) > 0 {
			execution := resp.Executions[0]
			workflowID := execution.Execution.WorkflowId
			runType := "sync"
			if strings.HasPrefix(workflowID, "clear-destination-") {
				runType = "clear"
			}
			jobResp.LastRunTime = execution.StartTime.AsTime().Format(time.RFC3339)
			jobResp.LastRunState = execution.Status.String()
			jobResp.LastRunType = runType
		}
	}
	return jobResp
}

func (s *JobService) getOrCreateSource(config dto.JobSourceConfig, projectID string, userID *int) (*models.Source, error) {
	// TODO: need to make source creation for job and source api at same place
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

func (s *JobService) getOrCreateDestination(config dto.JobDestinationConfig, projectID string, userID *int) (*models.Destination, error) {
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

func (s *JobService) GetClearDestinationStatus(ctx context.Context, projectID string, jobID int) (bool, error) {
	if _, err := s.jobORM.GetByID(jobID, true); err != nil {
		return false, fmt.Errorf("job not found id %d: %s", jobID, err)
	}
	return isClearRunning(ctx, s.tempClient, projectID, jobID)
}

func (s *JobService) ClearDestination(ctx context.Context, projectID string, jobID int, streamsConfig string) (map[string]interface{}, error) {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found id %d: %s", jobID, err)
	}

	if running, _ := isClearRunning(ctx, s.tempClient, projectID, jobID); running {
		return nil, fmt.Errorf("clear-destination is in progress: %w", constants.ErrInProgress)
	}

	// cancel sync if running
	if running, _ := isSyncRunning(ctx, s.tempClient, projectID, jobID); running {
		return nil, fmt.Errorf("sync is in progress, please cancel it before running clear-destination: %w", constants.ErrInProgress)
	}

	result, err := s.tempClient.ClearDestination(ctx, job, streamsConfig)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *JobService) CheckUniqueJobName(projectID string, jobName string) (bool, error) {
	return s.jobORM.IsJobNameUnique(projectID, jobName)
}

// worker services
func (s *JobService) UpdateSyncTelemetry(ctx context.Context, jobID int, workflowID string, event string) error {
	switch strings.ToLower(event) {
	case "started":
		telemetry.TrackSyncStart(ctx, jobID, workflowID)
	case "completed":
		telemetry.TrackSyncCompleted(ctx, jobID, workflowID)
	case "failed":
		telemetry.TrackSyncFailed(ctx, jobID, workflowID)
	}

	return nil
}
