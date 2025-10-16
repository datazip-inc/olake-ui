package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/models/dto"
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
	tempClient, err := temporal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client - error=%v", err)
	}
	return &JobService{
		jobORM:     database.NewJobORM(),
		sourceORM:  database.NewSourceORM(),
		destORM:    database.NewDestinationORM(),
		tempClient: tempClient,
	}, nil
}

func (s *JobService) GetAllJobs(projectID string) ([]dto.JobResponse, error) {
	jobs, err := s.jobORM.GetAllJobsByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs - project_id=%s error=%v", projectID, err)
	}

	jobResponses := make([]dto.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		jobResp, err := s.buildJobResponse(job, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to build job response - project_id=%s job_id=%d error=%v", projectID, job.ID, err)
		}
		jobResponses = append(jobResponses, jobResp)
	}

	return jobResponses, nil
}

func (s *JobService) CreateJob(ctx context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error {
	source, err := s.getOrCreateSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source - project_id=%s job_name=%s error=%v", projectID, req.Name, err)
	}

	dest, err := s.getOrCreateDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination - project_id=%s job_name=%s error=%v", projectID, req.Name, err)
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

	user := &models.User{ID: *userID}
	job.CreatedBy = user
	job.UpdatedBy = user

	if err := s.jobORM.Create(job); err != nil {
		return fmt.Errorf("failed to create job - project_id=%s job_name=%s source_id=%d destination_id=%d user_id=%v error=%v",
			projectID, req.Name, source.ID, dest.ID, userID, err)
	}

	_, err = s.tempClient.ManageSync(ctx, job.ProjectID, job.ID, job.Frequency, temporal.ActionCreate)
	if err != nil {
		return fmt.Errorf("failed to create temporal workflow - project_id=%s job_id=%d job_name=%s error=%v",
			projectID, job.ID, req.Name, err)
	}

	telemetry.TrackJobCreation(ctx, &models.Job{Name: req.Name})
	return nil
}

func (s *JobService) UpdateJob(ctx context.Context, req *dto.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	existingJob, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to find job for update - project_id=%s job_id=%d error=%v", projectID, jobID, err)
	}

	source, err := s.getOrCreateSource(req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source for job update - project_id=%s job_id=%d error=%v",
			projectID, jobID, err)
	}

	dest, err := s.getOrCreateDestination(req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination for job update - project_id=%s job_id=%d error=%v",
			projectID, jobID, err)
	}

	existingJob.Name = req.Name
	existingJob.SourceID = source
	existingJob.DestID = dest
	existingJob.Active = req.Activate
	existingJob.Frequency = req.Frequency
	existingJob.StreamsConfig = req.StreamsConfig
	existingJob.ProjectID = projectID

	user := &models.User{ID: *userID}
	existingJob.UpdatedBy = user

	if err := s.jobORM.Update(existingJob); err != nil {
		return fmt.Errorf("failed to update job - project_id=%s job_id=%d job_name=%s error=%v",
			projectID, jobID, req.Name, err)
	}

	_, err = s.tempClient.ManageSync(ctx, existingJob.ProjectID, existingJob.ID, existingJob.Frequency, temporal.ActionUpdate)
	if err != nil {
		return fmt.Errorf("failed to update temporal workflow - project_id=%s job_id=%d error=%v",
			projectID, existingJob.ID, err)
	}

	telemetry.TrackJobEntity(ctx)
	return nil
}

func (s *JobService) DeleteJob(ctx context.Context, jobID int) (string, error) {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return "", fmt.Errorf("failed to find job for deletion - job_id=%d error=%v", jobID, err)
	}

	jobName := job.Name

	_, err = s.tempClient.ManageSync(ctx, job.ProjectID, job.ID, job.Frequency, temporal.ActionDelete)
	if err != nil {
		return "", fmt.Errorf("failed to delete temporal workflow - project_id=%s job_id=%d error=%v",
			job.ProjectID, job.ID, err)
	}

	if err := s.jobORM.Delete(jobID); err != nil {
		return "", fmt.Errorf("failed to delete job - job_id=%d job_name=%s error=%v", jobID, jobName, err)
	}

	telemetry.TrackJobEntity(ctx)
	return jobName, nil
}

func (s *JobService) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job for sync - project_id=%s job_id=%d error=%v", projectID, jobID, err)
	}

	resp, err := s.tempClient.ManageSync(ctx, job.ProjectID, job.ID, job.Frequency, temporal.ActionTrigger)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger sync - project_id=%s job_id=%d error=%v", projectID, jobID, err)
	}
	return resp, nil
}

func (s *JobService) CancelJobRun(_ context.Context, projectID string, jobID int) (map[string]any, error) {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job for cancel - project_id=%s job_id=%d error=%v", projectID, jobID, err)
	}

	jobSlice := []*models.Job{job}
	if err := cancelAllJobWorkflows(s.tempClient, jobSlice, projectID); err != nil {
		return nil, fmt.Errorf("failed to cancel job workflow - project_id=%s job_id=%d error=%v", projectID, jobID, err)
	}
	// TODO : remove nested parsing from frontend
	return map[string]any{
		"message": "job workflow cancel requested successfully",
	}, nil
}

func (s *JobService) ActivateJob(_ context.Context, jobID int, req dto.JobStatusRequest, userID *int) error {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return fmt.Errorf("failed to find job for activation - job_id=%d error=%v", jobID, err)
	}

	job.Active = req.Activate

	user := &models.User{ID: *userID}
	job.UpdatedBy = user

	if err := s.jobORM.Update(job); err != nil {
		return fmt.Errorf("failed to update job activation status - job_id=%d activate=%v error=%v", jobID, req.Activate, err)
	}

	return nil
}

func (s *JobService) IsJobNameUnique(_ context.Context, projectID string, req dto.CheckUniqueJobNameRequest) (bool, error) {
	unique, err := s.jobORM.IsJobNameUnique(projectID, req.JobName)
	if err != nil {
		return false, fmt.Errorf("failed to check job name uniqueness - project_id=%s job_name=%s error=%v",
			projectID, req.JobName, err)
	}

	return unique, nil
}

func (s *JobService) GetJobTasks(ctx context.Context, projectID string, jobID int) ([]dto.JobTask, error) {
	job, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job for tasks - project_id=%s job_id=%d error=%v", projectID, jobID, err)
	}

	var tasks []dto.JobTask
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectID, job.ID, projectID, job.ID)

	resp, err := s.tempClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows - project_id=%s job_id=%d error=%v", projectID, jobID, err)
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

func (s *JobService) GetTaskLogs(_ context.Context, jobID int, filePath string) ([]map[string]interface{}, error) {
	_, err := s.jobORM.GetByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find job for logs - job_id=%d error=%v", jobID, err)
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))
	// Read the log file

	// Get home directory
	homeDir := docker.GetDefaultConfigDir()
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	logs, err := utils.ReadLogs(mainSyncDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs - job_id=%d path=%s error=%v", jobID, mainSyncDir, err)
	}
	// TODO: need to add activity logs as well with sync logs
	return logs, nil
}

func (s *JobService) buildJobResponse(job *models.Job, projectID string) (dto.JobResponse, error) {
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
			Name:    job.SourceID.Name,
			Type:    job.SourceID.Type,
			Config:  job.SourceID.Config,
			Version: job.SourceID.Version,
		}
	}

	if job.DestID != nil {
		jobResp.Destination = dto.DriverConfig{
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
	resp, err := s.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query:    query,
		PageSize: 1,
	})
	if err != nil {
		return dto.JobResponse{}, fmt.Errorf("failed to list workflows - project_id=%s job_id=%d error=%v", projectID, job.ID, err)
	}
	if len(resp.Executions) > 0 {
		jobResp.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
		jobResp.LastRunState = resp.Executions[0].Status.String()
	}

	return jobResp, nil
}

func (s *JobService) getOrCreateSource(config *dto.DriverConfig, projectID string, userID *int) (*models.Source, error) {
	sources, err := s.sourceORM.GetByNameAndType(config.Name, config.Type, projectID)
	if err == nil && len(sources) > 0 {
		source := sources[0]
		source.Config = config.Config
		source.Version = config.Version
		if userID != nil {
			source.UpdatedBy = &models.User{ID: *userID}
		}
		if err := s.sourceORM.Update(source); err != nil {
			return nil, fmt.Errorf("failed to update existing source - project_id=%s source_name=%s source_type=%s error=%v",
				projectID, config.Name, config.Type, err)
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
		return nil, fmt.Errorf("failed to create source - project_id=%s source_name=%s source_type=%s error=%v",
			projectID, config.Name, config.Type, err)
	}
	return source, nil
}

func (s *JobService) getOrCreateDestination(config *dto.DriverConfig, projectID string, userID *int) (*models.Destination, error) {
	destinations, err := s.destORM.GetByNameAndType(config.Name, config.Type, projectID)
	if err == nil && len(destinations) > 0 {
		dest := destinations[0]
		dest.Config = config.Config
		dest.Version = config.Version
		if userID != nil {
			dest.UpdatedBy = &models.User{ID: *userID}
		}
		if err := s.destORM.Update(dest); err != nil {
			return nil, fmt.Errorf("failed to update existing destination - project_id=%s destination_name=%s destination_type=%s error=%v",
				projectID, config.Name, config.Type, err)
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
		return nil, fmt.Errorf("failed to create destination - project_id=%s destination_name=%s destination_type=%s error=%v",
			projectID, config.Name, config.Type, err)
	}
	return dest, nil
}
