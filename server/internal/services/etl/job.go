package etl

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/types"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
	workflowservice "go.temporal.io/api/workflowservice/v1"
)

// Job-related methods on AppService

func (s Service) ListJobs(ctx context.Context, projectID string) ([]dto.JobResponse, error) {
	jobs, err := s.db.ListJobsByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %s", err)
	}

	lastRunByJobID, err := fetchLatestJobRunsByJobIDs(ctx, s.temporal, projectID, jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest job runs from temporal: %s", err)
	}

	jobResponses := make([]dto.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		var lastRun *JobLastRunInfo
		if lr, ok := lastRunByJobID[job.ID]; ok {
			lastRun = &lr
		}

		jobResp, err := s.buildJobResponse(job, lastRun, false)
		if err != nil {
			return nil, fmt.Errorf("failed to build job response: %s", err)
		}

		jobResponses = append(jobResponses, jobResp)
	}

	return jobResponses, nil
}

func (s Service) GetJob(ctx context.Context, projectID string, jobID int) (*dto.JobResponse, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return nil, fmt.Errorf("failed to get job: %s", err)
	}

	lastRunByJobID, err := fetchLatestJobRunsByJobIDs(ctx, s.temporal, projectID, []*models.Job{job})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest job runs from temporal: %s", err)
	}

	// It is valid for a job to have no previous runs; in that case we build the
	// response without last run information (same behavior as in ListJobs).
	var lastRun *JobLastRunInfo
	if lr, ok := lastRunByJobID[job.ID]; ok {
		lastRun = &lr
	}

	jobResponse, err := s.buildJobResponse(job, lastRun, true)
	if err != nil {
		return nil, fmt.Errorf("failed to build job response: %s", err)
	}

	return &jobResponse, nil
}

func (s Service) CreateJob(ctx context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error {
	unique, err := s.db.IsJobNameUniqueInProject(ctx, projectID, req.Name)
	if err != nil {
		return fmt.Errorf("failed to check job name uniqueness: %s", err)
	}
	if !unique {
		return fmt.Errorf("job name '%s' is not unique", req.Name)
	}

	source, err := s.upsertSource(ctx, req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source: %s", err)
	}

	if err := validateStreamsFormat(req.StreamsConfig, req.AvailableStreamsConfig, req.SelectedStreamsConfig, source.Version, nil); err != nil {
		return err
	}

	dest, err := s.upsertDestination(ctx, req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination: %s", err)
	}

	user := &models.User{ID: *userID}

	var advancedSettings *string
	if req.AdvancedSettings != nil {
		b, err := json.Marshal(req.AdvancedSettings)
		if err != nil {
			return fmt.Errorf("failed to serialise advanced_settings: %s", err)
		}
		s := string(b)
		advancedSettings = &s
	}

	job := &models.Job{
		Name:                   req.Name,
		SourceID:               source.ID,
		DestID:                 dest.ID,
		Source:                 source,
		Destination:            dest,
		Active:                 true,
		Frequency:              req.Frequency,
		StreamsConfig:          utils.StringPtr(req.StreamsConfig),
		AvailableStreamsConfig: utils.StringPtr(req.AvailableStreamsConfig),
		SelectedStreamsConfig:  utils.StringPtr(req.SelectedStreamsConfig),
		State:                  "{}",
		AdvancedSettings:       advancedSettings,
		ProjectID:              projectID,
		CreatedByID:            user.ID,
		UpdatedByID:            user.ID,
		CreatedBy:              user,
		UpdatedBy:              user,
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

	telemetry.TrackJobCreation(ctx, job)
	return nil
}

func (s Service) UpdateJob(ctx context.Context, req *dto.UpdateJobRequest, projectID string, jobID int, userID *int) error {
	// TODO: remove fetching existing job from database to verify it's existence, fetch only if the details aren't already available in the params/request. If job not exists it will fail during query execution.
	existingJob, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return fmt.Errorf("failed to get job: %s", err)
	}

	// Block when clear-destination is running
	clearRunning, _, err := isWorkflowRunning(ctx, s.temporal, projectID, jobID, temporal.ClearDestination)
	if err != nil {
		return fmt.Errorf("failed to check if clear-destination is running: %s", err)
	}
	if clearRunning {
		return fmt.Errorf("clear-destination is in progress, cannot update job")
	}

	source, err := s.upsertSource(ctx, req.Source, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process source for job update: %s", err)
	}

	if err := validateStreamsFormat(req.StreamsConfig, req.AvailableStreamsConfig, req.SelectedStreamsConfig, source.Version, existingJob); err != nil {
		return err
	}

	// Cancel sync before updating the job
	if err := cancelAllJobWorkflows(ctx, s.temporal, []*models.Job{existingJob}, projectID); err != nil {
		return fmt.Errorf("failed to cancel sync: %s", err)
	}

	// Handle stream difference if provided
	if req.DifferenceStreams != "" {
		var diffCatalog map[string]interface{}
		if err := json.Unmarshal([]byte(req.DifferenceStreams), &diffCatalog); err != nil {
			return fmt.Errorf("invalid difference_streams JSON: %s", err)
		}
		if len(diffCatalog) > 0 {
			if err := s.ClearDestination(ctx, projectID, jobID, req.DifferenceStreams, constants.DefaultCancelSyncWaitTime, false); err != nil {
				return fmt.Errorf("failed to run clear destination workflow: %s", err)
			}
			logger.Infof("successfully triggered clear destination workflow for job %d", existingJob.ID)
		}
	}

	dest, err := s.upsertDestination(ctx, req.Destination, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to process destination for job update: %s", err)
	}

	updateParams := map[string]any{
		"name":                     req.Name,
		"source_id":                source.ID,
		"dest_id":                  dest.ID,
		"active":                   req.Activate,
		"frequency":                req.Frequency,
		"streams_config":           utils.StringPtr(req.StreamsConfig),
		"project_id":               projectID,
		"updated_by_id":            *userID,
		"available_streams_config": utils.StringPtr(req.AvailableStreamsConfig),
		"selected_streams_config":  utils.StringPtr(req.SelectedStreamsConfig),
	}
	if req.AdvancedSettings != nil {
		b, err := json.Marshal(req.AdvancedSettings)
		if err != nil {
			return fmt.Errorf("failed to serialise advanced_settings: %s", err)
		}
		updateParams["advanced_settings"] = string(b)
	} else {
		updateParams["advanced_settings"] = nil
	}

	if err := s.db.UpdateJob(existingJob.ID, updateParams); err != nil {
		return fmt.Errorf("failed to update job: %s", err)
	}

	// TODO: check compensation/outbox pattern or any relevant measures to handle failure of update schedule workflow
	// Update temporal schedule only if frequency has changed
	if req.Frequency != existingJob.Frequency {
		err = s.temporal.UpdateSchedule(ctx, req.Frequency, projectID, existingJob.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to update temporal workflow: %s", err)
		}
	}

	return nil
}

func (s Service) DeleteJob(ctx context.Context, jobID int) (string, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return "", fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return "", fmt.Errorf("failed to find job: %s", err)
	}

	if err = s.temporal.DeleteSchedule(ctx, job.ProjectID, job.ID); err != nil {
		return "", fmt.Errorf("failed to delete temporal workflow: %s", err)
	}

	if err := s.db.DeleteJob(jobID); err != nil {
		return "", fmt.Errorf("failed to delete job: %s", err)
	}

	return job.Name, nil
}

func (s Service) SyncJob(ctx context.Context, projectID string, jobID int) (interface{}, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
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

func (s Service) CancelJobRun(ctx context.Context, projectID string, jobID int) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return fmt.Errorf("failed to find job: %s", err)
	}

	jobSlice := []*models.Job{job}
	if err := cancelAllJobWorkflows(ctx, s.temporal, jobSlice, projectID); err != nil {
		return fmt.Errorf("failed to cancel job workflow: %s", err)
	}
	return nil
}

func (s Service) ActivateJob(ctx context.Context, jobID int, req dto.JobStatusRequest, userID *int) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return fmt.Errorf("failed to find job: %s", err)
	}

	if req.Activate == job.Active {
		return nil
	}

	if req.Activate {
		if err := s.temporal.ResumeSchedule(ctx, job.ProjectID, job.ID); err != nil {
			return fmt.Errorf("failed to unpause schedule: %s", err)
		}
	} else {
		if err := s.temporal.PauseSchedule(ctx, job.ProjectID, job.ID); err != nil {
			return fmt.Errorf("failed to pause schedule: %s", err)
		}
	}

	updateParams := map[string]any{
		"active":        req.Activate,
		"updated_by_id": *userID,
	}

	if err := s.db.UpdateJob(job.ID, updateParams); err != nil {
		return fmt.Errorf("failed to update job activation status: %s", err)
	}

	return nil
}

func (s Service) ClearDestination(ctx context.Context, projectID string, jobID int, streamsConfig string, syncWaitTime time.Duration, resetState bool) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return fmt.Errorf("job not found: %s", err)
	}

	if job.Source == nil {
		return fmt.Errorf("job source details not found")
	}
	if err := utils.CheckClearDestinationCompatibility(job.Source.Version); err != nil {
		return err
	}

	if !job.Active {
		return fmt.Errorf("job is paused, please unpause to run clear destination")
	}

	// Pause the schedule to prevent a race condition where a new sync could start
	// after the current sync stops but before clear-destination executes.
	// The schedule will be automatically resumed by the worker after clear-destination completes successfully.
	if err := s.temporal.PauseSchedule(ctx, projectID, jobID); err != nil {
		return fmt.Errorf("failed to pause schedule: %s", err)
	}

	// Check if sync is running and wait for it to stop
	if err := waitForSyncToStop(ctx, s.temporal, projectID, jobID, syncWaitTime); err != nil {
		if rerr := s.temporal.ResumeSchedule(ctx, projectID, jobID); rerr != nil {
			return fmt.Errorf("wait error: %s, resume error: %s", err, rerr)
		}
		return fmt.Errorf("failed to wait for sync to stop: %s", err)
	}

	// for manual clear-destination, update the state file to empty object
	if resetState {
		if err := s.UpdateStateFile(jobID, "{}"); err != nil {
			return fmt.Errorf("failed to update state file: %s", err)
		}
		logger.Infof("state file updated to {} for manual clear-destination for job_id[%d]", jobID)
	}

	logger.Infof("running clear destination workflow for job %d for the following streams:\n%s", job.ID, streamsConfig)

	if err := s.temporal.ClearDestination(ctx, job, streamsConfig); err != nil {
		if rerr := s.temporal.ResumeSchedule(ctx, projectID, jobID); rerr != nil {
			return fmt.Errorf("clear destination error: %s, resume error: %s", err, rerr)
		}
		return fmt.Errorf("failed to clear destination: %s", err)
	}

	return nil
}

// GetStreamDifference diffs the job's stored catalog against the edited one, sent as
// updated_streams_config or updated_available_streams_config + updated_selected_streams_config. The
// two sides may be in different formats: a legacy job can be edited into the split format.
func (s Service) GetStreamDifference(ctx context.Context, _ string, jobID int, req dto.StreamDifferenceRequest) (map[string]interface{}, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", err)
	}
	if job.Source == nil {
		return nil, fmt.Errorf("job source details not found")
	}
	if err := utils.CheckClearDestinationCompatibility(job.Source.Version); err != nil {
		return nil, err
	}

	newAvailable, newSelected := req.UpdatedAvailableStreamsConfig, req.UpdatedSelectedStreamsConfig
	if err := validateStreamsFormat(req.UpdatedStreamsConfig, newAvailable, newSelected, job.Source.Version, job); err != nil {
		return nil, err
	}

	edited := types.StreamsCatalog{Streams: req.UpdatedStreamsConfig, Available: newAvailable, Selected: newSelected}
	diffCatalog, err := s.temporal.GetStreamDifference(ctx, job, job.StreamsCatalog(), edited)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream difference: %s", err)
	}

	diffCatalogJSON, err := json.Marshal(diffCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream difference: %s", err)
	}

	logger.Infof("stream difference retrieved successfully for job %d\n%s", job.ID, string(diffCatalogJSON))
	return diffCatalog, nil
}

// validateStreamsFormat rejects a catalog that: is not exactly one format (streams, or available and
// selected together); uses the split format on a source that does not support it; or downgrades a
// split-catalog job to legacy.
func validateStreamsFormat(streamsConfig, available, selected, sourceVersion string, existingJob *models.Job) error {
	split := available != "" && selected != ""
	switch {
	case (available != "") != (selected != ""):
		return fmt.Errorf("%w: available_streams_config and selected_streams_config must be set together", constants.ErrStreamsFormat)
	case split && streamsConfig != "":
		return fmt.Errorf("%w: set either streams_config or available_streams_config + selected_streams_config, not both", constants.ErrStreamsFormat)
	case !split && streamsConfig == "":
		return fmt.Errorf("%w: either streams_config or available_streams_config + selected_streams_config is required", constants.ErrStreamsFormat)
	}
	if split && !utils.SupportsStreamsV2(sourceVersion) {
		return fmt.Errorf("%w: source version %s does not support the split streams format (minimum %s)",
			constants.ErrStreamsFormat, sourceVersion, constants.MinStreamsV2Version)
	}
	if !split && existingJob != nil && existingJob.StreamsCatalog().IsSplit() {
		return fmt.Errorf("%w: job uses the split streams format, set available_streams_config and selected_streams_config", constants.ErrStreamsFormat)
	}
	return nil
}

// ConvertLegacyJobs converts, one at a time, the legacy jobs whose source reads the split format:
// every such job at startup, or one source's jobs (sourceID) after a source upgrade.
// A job that fails stays legacy and runs as before.
func (s Service) ConvertLegacyJobs(ctx context.Context, sourceID *int) error {
	logger.Infof("Converting legacy jobs to the split format...")
	jobs, err := s.db.ListLegacyCatalogJobs(sourceID)
	if err != nil {
		return fmt.Errorf("streams conversion: failed to list legacy jobs: %s", err)
	}
	return s.convertLegacyJobs(ctx, jobs)
}

// convertLegacyJobs converts each legacy job to the split format, one at a time. The guarded
// write in SetStreamsV2Catalog skips a job that was saved or deleted meanwhile.
func (s Service) convertLegacyJobs(ctx context.Context, jobs []*models.Job) error {
	for _, job := range jobs {
		if ctx.Err() != nil {
			return fmt.Errorf("streams conversion: context cancelled: %s", ctx.Err())
		}
		if job.Source == nil {
			logger.Warnf("streams conversion: job_id[%d] has no source, skipping", job.ID)
			continue
		}
		if !utils.SupportsStreamsV2(job.Source.Version) {
			continue
		}
		legacyCatalog := utils.StringValue(job.StreamsConfig)
		available, selected, err := s.ConvertCatalog(ctx, job.Source.Type, job.Source.Version, legacyCatalog)
		if err != nil {
			logger.Warnf("streams conversion: job_id[%d] stays legacy: %s", job.ID, err)
			continue
		}
		updated, err := s.db.SetStreamsV2Catalog(job.ID, legacyCatalog, available, selected)
		if err != nil {
			logger.Warnf("streams conversion: job_id[%d] failed to store the split catalog: %s", job.ID, err)
			continue
		}
		if !updated {
			logger.Infof("streams conversion: job_id[%d] changed during conversion, skipping", job.ID)
			continue
		}
		logger.Infof("Job %d converted to the split format successfully", job.ID)
	}

	logger.Infof("supported legacy jobs converted to the split format")
	return nil
}

func (s Service) GetClearDestinationStatus(ctx context.Context, projectID string, jobID int) (bool, error) {
	_, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return false, fmt.Errorf("job not found: %s", err)
	}

	isClearRunning, _, err := isWorkflowRunning(ctx, s.temporal, projectID, jobID, temporal.ClearDestination)
	if err != nil {
		return false, fmt.Errorf("failed to check if clear destination is running: %s", err)
	}

	return isClearRunning, nil
}

func (s Service) CheckUniqueName(ctx context.Context, projectID string, req dto.CheckUniqueNameRequest) (bool, error) {
	var tableType constants.TableType
	switch req.EntityType {
	case "job":
		tableType = constants.JobTable
	case "source":
		tableType = constants.SourceTable
	case "destination":
		tableType = constants.DestinationTable
	default:
		return false, fmt.Errorf("invalid entity type: %s", req.EntityType)
	}

	unique, err := s.db.IsNameUniqueInProject(ctx, projectID, req.Name, tableType)
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness: %s", err)
	}

	return unique, nil
}

func (s Service) GetJobTasks(ctx context.Context, projectID string, jobID int) ([]dto.JobTask, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	var tasks []dto.JobTask
	query := fmt.Sprintf("WorkflowId BETWEEN 'sync-%s-%d-' AND 'sync-%s-%d-z'", projectID, job.ID, projectID, job.ID)

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

func (s Service) GetTaskLogs(ctx context.Context, jobID int, filePath string, cursor int64, limit int, direction string) (*dto.TaskLogsResponse, error) {
	_, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return nil, fmt.Errorf("failed to find job: %s", err)
	}

	// Get and validate base directory from file path
	mainSyncDir, err := utils.GetAndValidateLogBaseDir(ctx, filePath)
	if err != nil {
		return nil, err
	}

	logs, err := utils.ReadLogs(ctx, mainSyncDir, cursor, limit, direction)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %s", err)
	}
	// TODO: need to add activity logs as well with sync logs
	return logs, nil
}

// TODO: frontend needs to send source id and destination id
func (s Service) buildJobResponse(job *models.Job, lastRun *JobLastRunInfo, includeConfig bool) (dto.JobResponse, error) {
	jobResp := dto.JobResponse{
		ID:        job.ID,
		Name:      job.Name,
		Frequency: job.Frequency,
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
		UpdatedAt: job.UpdatedAt.Format(time.RFC3339),
		Activate:  job.Active,
	}

	if includeConfig {
		jobResp.StreamsConfig = utils.StringValue(job.StreamsConfig)
		jobResp.AvailableStreamsConfig = utils.StringValue(job.AvailableStreamsConfig)
		jobResp.SelectedStreamsConfig = utils.StringValue(job.SelectedStreamsConfig)
	}

	if job.Source != nil {
		jobResp.Source = dto.DriverConfig{
			ID:      &job.Source.ID,
			Name:    job.Source.Name,
			Type:    job.Source.Type,
			Version: job.Source.Version,
		}
		jobResp.Source.Config = utils.Ternary(includeConfig, job.Source.Config, "").(string)
	}

	if job.Destination != nil {
		jobResp.Destination = dto.DriverConfig{
			ID:      &job.Destination.ID,
			Name:    job.Destination.Name,
			Type:    job.Destination.DestType,
			Version: job.Destination.Version,
		}
		jobResp.Destination.Config = utils.Ternary(includeConfig, job.Destination.Config, "").(string)
	}

	if job.CreatedBy != nil {
		jobResp.CreatedBy = job.CreatedBy.Username
	}
	if job.UpdatedBy != nil {
		jobResp.UpdatedBy = job.UpdatedBy.Username
	}

	if lastRun != nil {
		jobResp.LastRunTime = lastRun.StartTime.Format(time.RFC3339)
		jobResp.LastRunState = lastRun.Status.String()
		jobResp.LastRunType = utils.Ternary(lastRun.OperationType == temporal.Sync, "sync", "clear").(string)
	}

	if job.AdvancedSettings != nil && *job.AdvancedSettings != "" && *job.AdvancedSettings != "{}" {
		var advSettings dto.AdvancedSettings
		if err := json.Unmarshal([]byte(*job.AdvancedSettings), &advSettings); err != nil {
			return dto.JobResponse{}, fmt.Errorf("failed to parse advanced_settings for job %d: %s", job.ID, err)
		}
		jobResp.AdvancedSettings = &advSettings
	}

	return jobResp, nil
}

func (s Service) upsertSource(ctx context.Context, config *dto.DriverConfig, projectID string, userID *int) (*models.Source, error) {
	if config == nil {
		return nil, fmt.Errorf("source config is required")
	}

	// If ID provided, use that source as-is without modifying it.
	if config.ID != nil {
		return s.db.GetSourceByID(*config.ID)
	}

	// Otherwise, create a new source.
	// check source name uniqueness in the project
	unique, err := s.db.IsNameUniqueInProject(ctx, projectID, config.Name, constants.SourceTable)
	if err != nil {
		return nil, fmt.Errorf("failed to check source name uniqueness: %s", err)
	}
	if !unique {
		return nil, fmt.Errorf("source name '%s' is not unique", config.Name)
	}

	user := &models.User{ID: *userID}

	newSource := &models.Source{
		Name:        config.Name,
		Type:        config.Type,
		Config:      config.Config,
		Version:     config.Version,
		ProjectID:   projectID,
		CreatedByID: user.ID,
		UpdatedByID: user.ID,
		CreatedBy:   user,
		UpdatedBy:   user,
	}
	if err := s.db.CreateSource(newSource); err != nil {
		return nil, fmt.Errorf("failed to create source: %s", err)
	}

	return newSource, nil
}

func (s Service) upsertDestination(ctx context.Context, config *dto.DriverConfig, projectID string, userID *int) (*models.Destination, error) {
	if config == nil {
		return nil, fmt.Errorf("destination config is required")
	}

	// If ID provided, use that destination as-is without modifying it.
	if config.ID != nil {
		return s.db.GetDestinationByID(*config.ID)
	}

	// Otherwise, create a new destination.
	// check destination name uniqueness in the project
	unique, err := s.db.IsNameUniqueInProject(ctx, projectID, config.Name, constants.DestinationTable)
	if err != nil {
		return nil, fmt.Errorf("failed to check destination name uniqueness: %s", err)
	}
	if !unique {
		return nil, fmt.Errorf("destination name '%s' is not unique", config.Name)
	}

	user := &models.User{ID: *userID}

	newDest := &models.Destination{
		Name:        config.Name,
		DestType:    config.Type,
		Config:      config.Config,
		Version:     config.Version,
		ProjectID:   projectID,
		CreatedByID: user.ID,
		UpdatedByID: user.ID,
		CreatedBy:   user,
		UpdatedBy:   user,
	}

	if err := s.db.CreateDestination(newDest); err != nil {
		return nil, fmt.Errorf("failed to create destination: %s", err)
	}

	return newDest, nil
}

// worker service
func (s Service) UpdateSyncTelemetry(_ context.Context, req dto.UpdateSyncTelemetryRequest) error {
	info := telemetry.SyncEventInfo{
		JobID:                req.JobID,
		WorkflowID:           req.WorkflowID,
		ExecutionEnvironment: req.Environment,
		Properties:           req.Properties,
	}
	switch strings.ToLower(req.Event) {
	case "started":
		telemetry.TrackSyncEvent(info, telemetry.EventSyncStarted)
	case "completed":
		telemetry.TrackSyncEvent(info, telemetry.EventSyncCompleted)
	case "failed":
		telemetry.TrackSyncEvent(info, telemetry.EventSyncFailed)
	case "cancelled":
		telemetry.TrackSyncEvent(info, telemetry.EventSyncCancelled)
	}

	return nil
}

// RecoverFromClearDestination cancels stuck clear-destination workflows and restores normal sync schedule
// This is an internal recovery API for when clear-destination gets stuck in infinite retry
func (s Service) RecoverFromClearDestination(ctx context.Context, projectID string, jobID int) error {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return fmt.Errorf("job not found: %s", err)
	}

	isClearRunning, executions, err := isWorkflowRunning(ctx, s.temporal, projectID, jobID, temporal.ClearDestination)
	if err != nil {
		logger.Warnf("failed to check clear-destination status: %s", err)
	}

	// cancel running clear-destination workflows
	if isClearRunning {
		logger.Infof("found %d running clear-destination workflow(s) for job %d, cancelling...", len(executions), jobID)
		for _, exec := range executions {
			if err := s.temporal.CancelWorkflow(ctx, exec.Execution.WorkflowId, exec.Execution.RunId); err != nil {
				logger.Errorf("failed to cancel clear-destination workflow %s: %s", exec.Execution.WorkflowId, err)
				continue
			}
			logger.Infof("cancelled clear-destination workflow %s", exec.Execution.WorkflowId)
		}
	} else {
		logger.Infof("no running clear-destination workflows found for job %d", jobID)
	}

	// restore schedule back to sync workflow
	if err := s.temporal.RestoreSyncSchedule(ctx, job); err != nil {
		return fmt.Errorf("failed to restore schedule to sync workflow: %s", err)
	}
	logger.Infof("restored schedule to sync workflow for job %d", jobID)

	// resume the schedule
	if err := s.temporal.ResumeSchedule(ctx, projectID, jobID); err != nil {
		return fmt.Errorf("failed to resume schedule: %s", err)
	}
	logger.Infof("resumed schedule for job %d", jobID)

	return nil
}

// StreamLogArchive creates and streams a tar.gz archive of job logs to the provided writer.
func (s Service) StreamLogArchive(ctx context.Context, jobID int, taskLogFilePath string, writer io.Writer) error {
	baseDir, err := utils.GetAndValidateLogBaseDir(ctx, taskLogFilePath)
	if err != nil {
		return err
	}

	_, err = utils.GetAndValidateSyncFolder(ctx, baseDir)
	if err != nil {
		return err
	}

	logger.Infof("Starting log archive creation for job_id[%d]", jobID)

	// Create streaming pipeline: tarWriter → gzipWriter → writer
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	logger.Debugf("Adding files from %s to archive", baseDir)
	if err := utils.AddFilesToArchive(ctx, baseDir, tarWriter); err != nil {
		return err
	}

	logger.Infof("Successfully created log archive for job_id[%d]", jobID)

	return nil
}

func (s Service) UpdateStateFile(jobID int, stateFile string) error {
	_, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		if errors.Is(err, constants.ErrJobNotFound) {
			return fmt.Errorf("%w: %v", constants.ErrJobNotFound, err)
		}
		return fmt.Errorf("job not found: %s", err)
	}

	if err := s.db.UpdateJob(jobID, map[string]any{"state": stateFile}); err != nil {
		return fmt.Errorf("failed to update job: %s", err)
	}

	logger.Infof("state file updated successfully for job_id[%d] with state: %s", jobID, stateFile)
	return nil
}
