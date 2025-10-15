package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/internal/temporal"
	"github.com/datazip/olake-ui/server/utils"
)

type SourceService struct {
	sourceORM  *database.SourceORM
	userORM    *database.UserORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func NewSourceService() (*SourceService, error) {
	tempClient, err := temporal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client - error=%v", err)
	}

	return &SourceService{
		sourceORM:  database.NewSourceORM(),
		userORM:    database.NewUserORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *SourceService) GetAllSources(_ context.Context, projectID string) ([]dto.SourceDataItem, error) {
	sources, err := s.sourceORM.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sources - project_id=%s error=%v", projectID, err)
	}

	sourceIDs := make([]int, 0, len(sources))
	for _, src := range sources {
		sourceIDs = append(sourceIDs, src.ID)
	}

	var allJobs []*models.Job
	allJobs, err = s.jobORM.GetBySourceID(sourceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs - project_id=%s source_count=%d error=%v", projectID, len(sourceIDs), err)
	}

	jobsBySourceID := make(map[int][]*models.Job)
	for _, job := range allJobs {
		if job.SourceID != nil {
			jobsBySourceID[job.SourceID.ID] = append(jobsBySourceID[job.SourceID.ID], job)
		}
	}

	items := make([]dto.SourceDataItem, 0, len(sources))
	for _, src := range sources {
		item := dto.SourceDataItem{
			ID:        src.ID,
			Name:      src.Name,
			Type:      src.Type,
			Version:   src.Version,
			Config:    src.Config,
			CreatedAt: src.CreatedAt.Format(time.RFC3339),
			UpdatedAt: src.UpdatedAt.Format(time.RFC3339),
		}
		setUsernames(&item.CreatedBy, &item.UpdatedBy, src.CreatedBy, src.UpdatedBy)

		jobs := jobsBySourceID[src.ID]
		jobItems, err := s.buildJobDataItems(jobs, projectID, "source")
		if err != nil {
			return nil, fmt.Errorf("failed to process job data items - project_id=%s source_id=%d error=%v", projectID, src.ID, err)
		}
		item.Jobs = jobItems

		items = append(items, item)
	}

	return items, nil
}

func (s *SourceService) CreateSource(ctx context.Context, req *dto.CreateSourceRequest, projectID string, userID *int) error {
	src := &models.Source{
		Name:      req.Name,
		Type:      req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}

	user := &models.User{ID: *userID}
	src.CreatedBy = user
	src.UpdatedBy = user

	if err := s.sourceORM.Create(src); err != nil {
		return fmt.Errorf("failed to create source - project_id=%s source_name=%s source_type=%s user_id=%v error=%v", projectID, req.Name, req.Type, userID, err)
	}

	telemetry.TrackSourceCreation(ctx, src)
	return nil
}

func (s *SourceService) UpdateSource(ctx context.Context, projectID string, id int, req *dto.UpdateSourceRequest, userID *int) error {
	existing, err := s.sourceORM.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to find source for update - project_id=%s source_id=%d error=%v: %w", projectID, id, err, constants.ErrSourceNotFound)
	}

	existing.Name = req.Name
	existing.Config = req.Config
	existing.Type = req.Type
	existing.Version = req.Version

	user := &models.User{ID: *userID}
	existing.UpdatedBy = user

	jobs, err := s.jobORM.GetBySourceID([]int{existing.ID})
	if err != nil {
		return fmt.Errorf("failed to fetch jobs for source update - project_id=%s source_id=%d error=%v", projectID, existing.ID, err)
	}

	if err := cancelAllJobWorkflows(s.tempClient, jobs, projectID); err != nil {
		return fmt.Errorf("failed to cancel workflows for source update - project_id=%s source_id=%d job_count=%d error=%v", projectID, existing.ID, len(jobs), err)
	}

	if err := s.sourceORM.Update(existing); err != nil {
		return fmt.Errorf("failed to update source - project_id=%s source_id=%d source_name=%s error=%v", projectID, id, req.Name, err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return nil
}

func (s *SourceService) DeleteSource(ctx context.Context, id int) (*dto.DeleteSourceResponse, error) {
	src, err := s.sourceORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find source for deletion - source_id=%d error=%v: %w", id, err, constants.ErrSourceNotFound)
	}

	jobs, err := s.jobORM.GetBySourceID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs for source deletion - source_id=%d error=%v", id, err)
	}

	jobIDs := make([]int, 0, len(jobs))
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	if err := s.jobORM.UpdateAllJobs(jobIDs); err != nil {
		return nil, fmt.Errorf("failed to update jobs for source deletion - source_id=%d job_count=%d error=%v", id, len(jobIDs), err)
	}

	if err := s.sourceORM.Delete(id); err != nil {
		return nil, fmt.Errorf("failed to delete source - source_id=%d source_name=%s error=%v", id, src.Name, err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return &dto.DeleteSourceResponse{Name: src.Name}, nil
}

func (s *SourceService) TestConnection(ctx context.Context, req *dto.SourceTestConnectionRequest) (map[string]interface{}, []map[string]interface{}, error) {
	if s.tempClient == nil {
		return nil, nil, fmt.Errorf("temporal client not available - source_type=%s source_version=%s", req.Type, req.Version)
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt config for test connection - source_type=%s source_version=%s error=%v", req.Type, req.Version, err)
	}
	workflowID := fmt.Sprintf("test-connection-%s-%d", req.Type, time.Now().Unix())
	result, err := s.tempClient.TestConnection(ctx, workflowID, "source", req.Type, req.Version, encryptedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connection test failed - source_type=%s source_version=%s error=%v", req.Type, req.Version, err)
	}
	homeDir := docker.GetDefaultConfigDir()
	mainLogDir := filepath.Join(homeDir, workflowID)
	logs, err := utils.ReadLogs(mainLogDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read logs for test connection - source_type=%s source_version=%s error=%v", req.Type, req.Version, err)
	}
	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}

	return result, logs, nil
}

func (s *SourceService) GetSourceCatalog(ctx context.Context, req *dto.StreamsRequest) (map[string]interface{}, error) {
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available - source_type=%s job_id=%d", req.Type, req.JobID)
	}

	oldStreams := ""
	if req.JobID >= 0 {
		job, err := s.jobORM.GetByID(req.JobID, true)
		if err != nil {
			return nil, fmt.Errorf("failed to find job for catalog - job_id=%d source_type=%s error=%v", req.JobID, req.Type, err)
		}
		oldStreams = job.StreamsConfig
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt config for catalog - source_type=%s job_id=%d error=%v", req.Type, req.JobID, err)
	}

	newStreams, err := s.tempClient.GetCatalog(
		ctx,
		req.Type,
		req.Version,
		encryptedConfig,
		oldStreams,
		req.JobName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog - source_type=%s source_version=%s job_id=%d error=%v", req.Type, req.Version, req.JobID, err)
	}

	return newStreams, nil
}

func (s *SourceService) GetSourceJobs(_ context.Context, id int) ([]*models.Job, error) {
	if _, err := s.sourceORM.GetByID(id); err != nil {
		return nil, fmt.Errorf("failed to find source - source_id=%d error=%v: %w", id, err, constants.ErrSourceNotFound)
	}

	jobs, err := s.jobORM.GetBySourceID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source - source_id=%d error=%v", id, err)
	}

	return jobs, nil
}

func (s *SourceService) GetSourceVersions(ctx context.Context, sourceType string) (map[string]interface{}, error) {
	if sourceType == "" {
		return nil, fmt.Errorf("source type is required")
	}

	imageName := fmt.Sprintf("olakego/source-%s", sourceType)
	versions, _, err := utils.GetDriverImageTags(ctx, imageName, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker versions - source_type=%s image=%s error=%v", sourceType, imageName, err)
	}

	return map[string]interface{}{"version": versions}, nil
}

// TODO: cache spec in db for each version
func (s *SourceService) GetSourceSpec(ctx context.Context, req *dto.SpecRequest) (dto.SpecResponse, error) {
	if err := dto.Validate(&req); err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to validate get spec request - source_type=%s source_version=%s error=%v", req.Type, req.Version, err)
	}

	if s.tempClient == nil {
		return dto.SpecResponse{}, fmt.Errorf("temporal client not available - source_type=%s source_version=%s", req.Type, req.Version)
	}

	specOut, err := s.tempClient.FetchSpec(ctx, "", req.Type, req.Version)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get spec - source_type=%s source_version=%s error=%v", req.Type, req.Version, err)
	}

	return dto.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOut.Spec,
	}, nil
}

// Helper methods
func (s *SourceService) buildJobDataItems(jobs []*models.Job, _, _ string) ([]dto.JobDataItem, error) {
	jobItems := make([]dto.JobDataItem, 0, len(jobs))
	for _, job := range jobs {
		item := dto.JobDataItem{
			ID:   job.ID,
			Name: job.Name,
		}
		jobItems = append(jobItems, item)
	}
	return jobItems, nil
}

func setUsernames(createdBy, updatedBy *string, createdByUser, updatedByUser *models.User) {
	if createdByUser != nil {
		*createdBy = createdByUser.Username
	}
	if updatedByUser != nil {
		*updatedBy = updatedByUser.Username
	}
}
