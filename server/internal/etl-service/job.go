package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/temporal"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
	"go.temporal.io/api/workflowservice/v1"
)

// Job-related methods on AppService

func (s *ETLService) GetAllJobs(ctx context.Context, projectID string) ([]dto.JobResponse, error) {
	jobs, err := s.db.ListJobsByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %s", err)
	}

	jobResponses := make([]dto.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		jobResp, err := s.buildJobResponse(ctx, job, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to build job response: %s", err)
		}
		jobResponses = append(jobResponses, jobResp)
	}

	return jobResponses, nil
}

func (s *ETLService) CreateJob(ctx context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error {
	source, err := s.upsertSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source: %s", err)
	}

	dest, err := s.upsertDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination: %s", err)
	}

	user := &models.User{ID: *userID}
	job := &models.Job{
		Name:          req.Name,
		SourceID:      source,
		DestID:        dest,
		Active:        true,
		Frequency:     req.Frequency,
		StreamsConfig: req.StreamsConfig,
		State:         "{}",
		ProjectID:     projectID,
		CreatedBy:     user,
		UpdatedBy:     user,
	}
	if err := s.db.CreateJob(job); err != nil {
		return fmt.Errorf("failed to create job: %s", err)
	}

	defer func() {
		if err != nil {
			if err := s.db.DeleteJob(job.ID); err != nil {
				logger.Errorf("failed to delete job: %s", err)
			}
		}
	}()

	if err = s.temporal.CreateSchedule(ctx, job); err != nil {
		return fmt.Errorf("failed to create temporal workflow: %s", err)
	}

	telemetry.TrackJobCreation(ctx, &models.Job{Name: req.Name})
	return nil
}

func (s *ETLService) UpdateJob(ctx context.Context, req *dto.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	existingJob, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to get job: %s", err)
	}

	// Block when clear-destination is running
	if clearRunning, _ := isWorkflowRunning(ctx, s.temporal, projectID, jobID, temporal.ClearDestination); clearRunning {
		return fmt.Errorf("clear-destination is in progress, cannot update job")
	}

	// Cancel sync before updating the job
	if syncRunning, _ := isWorkflowRunning(ctx, s.temporal, projectID, jobID, temporal.Sync); syncRunning {
		logger.Infof("sync is running for job %d, cancelling sync workflow", jobID)
		jobSlice := []*models.Job{existingJob}
		if err := cancelAllJobWorkflows(ctx, s.temporal, jobSlice, projectID); err != nil {
			return fmt.Errorf("failed to cancel sync: %s", err)
		}
		logger.Infof("successfully cancelled sync for job %d", jobID)
	}

	// Handle stream difference if provided
	if req.DifferenceStreams != "" {
		var diffCatalog map[string]interface{}
		if err := json.Unmarshal([]byte(req.DifferenceStreams), &diffCatalog); err != nil {
			return fmt.Errorf("invalid difference_streams JSON: %s", err)
		}

		if len(diffCatalog) > 0 {
			logger.Infof("stream difference detected for job %d, running clear destination workflow", existingJob.ID)
			if _, err := s.ClearDestination(ctx, projectID, jobID, req.DifferenceStreams); err != nil {
				return fmt.Errorf("failed to run clear destination workflow: %s", err)
			}
			logger.Infof("successfully triggered clear destination workflow for job %d", existingJob.ID)
		}
	}

	// Snapshot previous job state for compensation on schedule update failure
	prevJob := *existingJob

	source, err := s.upsertSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source for job update: %s", err)
	}

	dest, err := s.upsertDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination for job update: %s", err)
	}

	existingJob.Name = req.Name // TODO: job name cant be changed
	existingJob.SourceID = source
	existingJob.DestID = dest
	existingJob.Active = req.Activate
	existingJob.Frequency = req.Frequency
	existingJob.StreamsConfig = req.StreamsConfig
	existingJob.ProjectID = projectID
	existingJob.UpdatedBy = &models.User{ID: *userID}

	if err := s.db.UpdateJob(existingJob); err != nil {
		return fmt.Errorf("failed to update job: %s", err)
	}

	err = s.temporal.UpdateScheduleSpec(ctx, existingJob.Frequency, existingJob.ProjectID, existingJob.ID)
	if err != nil {
		// Compensation: restore previous DB state if schedule update fails
		if rerr := s.db.UpdateJob(&prevJob); rerr != nil {
			logger.Errorf("failed to restore job after schedule update error: %s", rerr)
		}
		return fmt.Errorf("failed to update temporal workflow: %s", err)
	}

	telemetry.TrackJobEntity(ctx)
	return nil
}

func (s *ETLService) DeleteJob(ctx context.Context, jobID int) (string, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return "", fmt.Errorf("failed to find job: %s", err)
	}

	if err = s.temporal.DeleteSchedule(ctx, job.ProjectID, job.ID); err != nil {
		return "", fmt.Errorf("failed to delete temporal workflow: %s", err)
	}

	if err := s.db.DeleteJob(jobID); err != nil {
		return "", fmt.Errorf("failed to delete job: %s", err)
	}

	telemetry.TrackJobEntity(ctx)
	return job.Name, nil
}

func (s *ETLService) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	if !job.Active {
		return nil, fmt.Errorf("job is paused, please unpause to run sync")
	}

	if err := s.temporal.TriggerSchedule(ctx, projectID, jobID); err != nil {
		return nil, fmt.Errorf("failed to trigger sync: %s", err)
	}

	return map[string]any{
		"message": "sync triggered successfully",
	}, nil
}

func (s *ETLService) CancelJobRun(ctx context.Context, projectID string, jobID int) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to find job: %s", err)
	}

	jobSlice := []*models.Job{job}
	if err := cancelAllJobWorkflows(ctx, s.temporal, jobSlice, projectID); err != nil {
		return fmt.Errorf("failed to cancel job workflow: %s", err)
	}
	return nil
}

func (s *ETLService) ActivateJob(ctx context.Context, jobID int, req dto.JobStatusRequest, userID *int) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to find job: %s", err)
	}

	if req.Activate == job.Active {
		return nil
	}

	if req.Activate {
		if err := s.temporal.UnpauseSchedule(ctx, job.ProjectID, job.ID); err != nil {
			return fmt.Errorf("failed to unpause schedule: %s", err)
		}
	} else {
		if err := s.temporal.PauseSchedule(ctx, job.ProjectID, job.ID); err != nil {
			return fmt.Errorf("failed to pause schedule: %s", err)
		}
	}

	job.Active = req.Activate
	user := &models.User{ID: *userID}
	job.UpdatedBy = user

	if err := s.db.UpdateJob(job); err != nil {
		return fmt.Errorf("failed to update job activation status: %s", err)
	}

	return nil
}

func (s *ETLService) ClearDestination(ctx context.Context, projectID string, jobID int, streamsConfig string) (map[string]interface{}, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", err)
	}

	if !job.Active {
		return nil, fmt.Errorf("job is paused, please unpause to run clear destination")
	}

	// Check if sync is running and wait for it to stop
	if running, _ := isWorkflowRunning(ctx, s.temporal, projectID, jobID, temporal.Sync); running {
		if err := waitForSyncToStop(ctx, s.temporal, projectID, jobID, 5*time.Second); err != nil {
			return nil, fmt.Errorf("sync is in progress, please cancel it before running clear-destination")
		}
	}

	if err := s.temporal.ClearDestination(ctx, job, streamsConfig); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "Clear destination initiated successfully",
	}, nil
}

func (s *ETLService) GetStreamDifference(ctx context.Context, projectID string, jobID int, req dto.StreamDifferenceRequest) (map[string]interface{}, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", err)
	}

	diffCatalog, err := s.temporal.GetDifferenceStreams(ctx, job, job.StreamsConfig, req.UpdatedStreamsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream difference: %s", err)
	}

	return diffCatalog, nil
}

func (s *ETLService) IsJobNameUnique(_ context.Context, projectID string, req dto.CheckUniqueJobNameRequest) (bool, error) {
	unique, err := s.db.IsJobNameUniqueInProject(projectID, req.JobName)
	if err != nil {
		return false, fmt.Errorf("failed to check job name uniqueness: %s", err)
	}

	return unique, nil
}

func (s *ETLService) GetJobTasks(ctx context.Context, projectID string, jobID int) ([]dto.JobTask, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	var tasks []dto.JobTask
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)

	resp, err := s.temporal.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
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

		opType := syncWorkflowOperationType(execution)
		jobType := utils.Ternary(opType == temporal.Sync, "sync", "clear").(string)
		tasks = append(tasks, dto.JobTask{
			Runtime:   runTime,
			StartTime: startTime.Format(time.RFC3339),
			Status:    execution.Status.String(),
			FilePath:  execution.Execution.WorkflowId,
			JobType:   jobType,
		})
	}

	return tasks, nil
}

func (s *ETLService) GetTaskLogs(_ context.Context, jobID int, filePath string) ([]map[string]interface{}, error) {
	_, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))

	// Get home directory
	homeDir := constants.DefaultConfigDir
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	logs, err := utils.ReadLogs(mainSyncDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %s", err)
	}
	// TODO: need to add activity logs as well with sync logs
	return logs, nil
}

// TODO: frontend needs to send source id and destination id
func (s *ETLService) buildJobResponse(ctx context.Context, job *models.Job, projectID string) (dto.JobResponse, error) {
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
		jobResp.Source = dto.DriverConfig{
			ID:      &job.SourceID.ID,
			Name:    job.SourceID.Name,
			Type:    job.SourceID.Type,
			Config:  job.SourceID.Config,
			Version: job.SourceID.Version,
		}
	}

	if job.DestID != nil {
		jobResp.Destination = dto.DriverConfig{
			ID:      &job.DestID.ID,
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

	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)
	resp, err := s.temporal.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query:    query,
		PageSize: 1,
	})
	if err != nil {
		return dto.JobResponse{}, fmt.Errorf("failed to list workflows: %s", err)
	}
	if len(resp.Executions) > 0 {
		opType := syncWorkflowOperationType(resp.Executions[0])
		jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
		jobResp.LastRunState = resp.Executions[0].Status.String()
		jobResp.LastRunType = utils.Ternary(opType == temporal.Sync, "sync", "clear").(string)
	}

	return jobResp, nil
}

func (s *ETLService) upsertSource(config *dto.DriverConfig, projectID string, userID *int) (*models.Source, error) {
	if config == nil {
		return nil, fmt.Errorf("source config is required")
	}

	// If ID provided, use that source as-is without modifying it.
	if config.ID != nil {
		return s.db.GetSourceByID(*config.ID)
	}

	user := &models.User{ID: *userID}
	// Otherwise, create a new source.
	newSource := &models.Source{
		Name:      config.Name,
		Type:      config.Type,
		Config:    config.Config,
		Version:   config.Version,
		ProjectID: projectID,
		CreatedBy: user,
		UpdatedBy: user,
	}
	if err := s.db.CreateSource(newSource); err != nil {
		return nil, fmt.Errorf("failed to create source: %s", err)
	}

	return newSource, nil
}

func (s *ETLService) upsertDestination(config *dto.DriverConfig, projectID string, userID *int) (*models.Destination, error) {
	if config == nil {
		return nil, fmt.Errorf("destination config is required")
	}

	// If ID provided, use that destination as-is without modifying it.
	if config.ID != nil {
		return s.db.GetDestinationByID(*config.ID)
	}

	user := &models.User{ID: *userID}
	// Otherwise, create a new destination.
	newDest := &models.Destination{
		Name:      config.Name,
		DestType:  config.Type,
		Config:    config.Config,
		Version:   config.Version,
		ProjectID: projectID,
		CreatedBy: user,
		UpdatedBy: user,
	}

	if err := s.db.CreateDestination(newDest); err != nil {
		return nil, fmt.Errorf("failed to create destination: %s", err)
	}

	return newDest, nil
}

// worker service
func (s *ETLService) UpdateSyncTelemetry(ctx context.Context, jobID int, workflowID, event string) error {
	switch strings.ToLower(event) {
	case "started":
		telemetry.TrackSyncStart(ctx, jobID, workflowID)
	case "completed":
		telemetry.TrackSyncCompleted(jobID, workflowID)
	case "failed":
		telemetry.TrackSyncFailed(jobID, workflowID)
	}

	return nil
}
