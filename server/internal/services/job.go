package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
	"go.temporal.io/api/workflowservice/v1"
)

// Job-related methods on AppService

func (s *AppService) GetAllJobs(ctx context.Context, projectID string) ([]dto.JobResponse, error) {
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

func (s *AppService) CreateJob(ctx context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error {
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

	if err = s.temporal.CreateSchedule(ctx, job.Frequency, job.ProjectID, job.ID); err != nil {
		return fmt.Errorf("failed to create temporal workflow: %s", err)
	}

	telemetry.TrackJobCreation(ctx, &models.Job{Name: req.Name})
	return nil
}

func (s *AppService) UpdateJob(ctx context.Context, req *dto.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	existingJob, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to get job: %s", err)
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

	err = s.temporal.UpdateSchedule(ctx, existingJob.Frequency, existingJob.ProjectID, existingJob.ID)
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

func (s *AppService) DeleteJob(ctx context.Context, jobID int) (string, error) {
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

func (s *AppService) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	if err := s.temporal.TriggerSchedule(ctx, job.ProjectID, job.ID); err != nil {
		return nil, fmt.Errorf("failed to trigger sync: %s", err)
	}

	return map[string]any{
		"message": "sync triggered successfully",
	}, nil
}

func (s *AppService) CancelJobRun(ctx context.Context, projectID string, jobID int) (map[string]any, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	jobSlice := []*models.Job{job}
	if err := cancelAllJobWorkflows(ctx, s.temporal, jobSlice, projectID); err != nil {
		return nil, fmt.Errorf("failed to cancel job workflow: %s", err)
	}
	// TODO : remove nested parsing from frontend
	return map[string]any{
		"message": "job workflow cancel requested successfully",
	}, nil
}

func (s *AppService) ActivateJob(_ context.Context, jobID int, req dto.JobStatusRequest, userID *int) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to find job: %s", err)
	}

	job.Active = req.Activate

	user := &models.User{ID: *userID}
	job.UpdatedBy = user

	if err := s.db.UpdateJob(job); err != nil {
		return fmt.Errorf("failed to update job activation status: %s", err)
	}

	return nil
}

func (s *AppService) IsJobNameUnique(_ context.Context, projectID string, req dto.CheckUniqueJobNameRequest) (bool, error) {
	unique, err := s.db.IsJobNameUniqueInProject(projectID, req.JobName)
	if err != nil {
		return false, fmt.Errorf("failed to check job name uniqueness: %s", err)
	}

	return unique, nil
}

func (s *AppService) GetJobTasks(ctx context.Context, projectID string, jobID int) ([]dto.JobTask, error) {
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
		tasks = append(tasks, dto.JobTask{
			Runtime:   runTime,
			StartTime: startTime.Format(time.RFC3339),
			Status:    execution.Status.String(),
			FilePath:  execution.Execution.WorkflowId,
		})
	}

	return tasks, nil
}

func (s *AppService) GetTaskLogs(_ context.Context, jobID int, filePath string) ([]map[string]interface{}, error) {
	_, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))
	// Read the log file

	// Get home directory
	homeDir := docker.GetDefaultConfigDir()
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	logs, err := utils.ReadLogs(mainSyncDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %s", err)
	}
	// TODO: need to add activity logs as well with sync logs
	return logs, nil
}

// TODO: frontend needs to send source id and destination id
func (s *AppService) buildJobResponse(ctx context.Context, job *models.Job, projectID string) (dto.JobResponse, error) {
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
		jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
		jobResp.LastRunState = resp.Executions[0].Status.String()
	}

	return jobResp, nil
}

func (s *AppService) upsertSource(config *dto.DriverConfig, projectID string, userID *int) (*models.Source, error) {
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

func (s *AppService) upsertDestination(config *dto.DriverConfig, projectID string, userID *int) (*models.Destination, error) {
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
