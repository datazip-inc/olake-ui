package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
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
	logs.Info("Creating source service")
	tempClient, err := temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %s", err)
		return nil, fmt.Errorf("%s temporal client: %s", constants.ErrFailedToCreate, err)
	}

	return &SourceService{
		sourceORM:  database.NewSourceORM(),
		userORM:    database.NewUserORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *SourceService) GetAllSources(ctx context.Context, projectID string) ([]dto.SourceDataItem, error) {
	logs.Info("Getting all sources for project: %s", projectID)
	sources, err := s.sourceORM.GetAll()
	if err != nil {
		return nil, fmt.Errorf("%s sources: %s", constants.ErrFailedToRetrieve, err)
	}

	sourceIDs := make([]int, 0, len(sources))
	for _, src := range sources {
		sourceIDs = append(sourceIDs, src.ID)
	}

	var allJobs []*models.Job
	if len(sourceIDs) > 0 {
		allJobs, err = s.jobORM.GetBySourceIDs(sourceIDs)
		if err != nil {
			return nil, fmt.Errorf("%s jobs: %s", constants.ErrFailedToRetrieve, err)
		}
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
			return nil, fmt.Errorf("%s job data items: %s", constants.ErrFailedToProcess, err)
		}
		item.Jobs = jobItems

		items = append(items, item)
	}

	return items, nil
}

func (s *SourceService) CreateSource(ctx context.Context, req dto.CreateSourceRequest, projectID string, userID *int) error {
	if err := dto.Validate(&req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	logs.Info("Creating source with projectID: %s", projectID)
	src := &models.Source{
		Name:      req.Name,
		Type:      req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}

	if userID != nil {
		// Align with other services by avoiding extra DB lookup
		user := &models.User{ID: *userID}
		src.CreatedBy = user
		src.UpdatedBy = user
	}

	if err := s.sourceORM.Create(src); err != nil {
		return fmt.Errorf("%s source: %s", constants.ErrFailedToCreate, err)
	}

	telemetry.TrackSourceCreation(ctx, src)
	return nil
}

func (s *SourceService) UpdateSource(ctx context.Context, projectID string, id int, req dto.UpdateSourceRequest, userID *int) error {
	if err := dto.Validate(&req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	logs.Info("Updating source with id: %d", id)
	existing, err := s.sourceORM.GetByID(id)
	if err != nil {
		return constants.ErrSourceNotFound
	}

	existing.Name = req.Name
	existing.Config = req.Config
	existing.Type = req.Type
	existing.Version = req.Version

	if userID != nil {
		user := &models.User{ID: *userID}
		existing.UpdatedBy = user
	}

	// Cancel workflows for jobs linked to this source before persisting change
	jobs, err := s.jobORM.GetBySourceID(existing.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs for source %d: %s", existing.ID, err)
	}
	for _, job := range jobs {
		if err := cancelJobWorkflow(s.tempClient, job, projectID); err != nil {
			return fmt.Errorf("failed to cancel workflow for job %d: %s", job.ID, err)
		}
	}

	if err := s.sourceORM.Update(existing); err != nil {
		return fmt.Errorf("%s source: %s", constants.ErrFailedToUpdate, err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return nil
}

func (s *SourceService) DeleteSource(ctx context.Context, id int) (*dto.DeleteSourceResponse, error) {
	logs.Info("Deleting source with id: %d", id)
	src, err := s.sourceORM.GetByID(id)
	if err != nil {
		return nil, constants.ErrSourceNotFound
	}

	jobs, err := s.jobORM.GetBySourceID(id)
	if err != nil {
		return nil, fmt.Errorf("%s jobs for source: %s", constants.ErrFailedToRetrieve, err)
	}

	jobIDs := make([]int, 0, len(jobs))
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	if len(jobIDs) > 0 {
		if err := s.jobORM.UpdateAllJobs(jobIDs); err != nil {
			return nil, fmt.Errorf("%s jobs: %s", constants.ErrFailedToUpdate, err)
		}
	}

	if err := s.sourceORM.Delete(id); err != nil {
		return nil, fmt.Errorf("%s source: %s", constants.ErrFailedToDelete, err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return &dto.DeleteSourceResponse{Name: src.Name}, nil
}

func (s *SourceService) TestConnection(ctx context.Context, req dto.SourceTestConnectionRequest) (map[string]interface{}, error) {
	if err := dto.Validate(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logs.Info("Testing connection for source: %v", req.Type)
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	if req.Type == "" || req.Version == "" {
		return nil, fmt.Errorf("source type and version are required")
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		logs.Error("Failed to encrypt config: %v", err)
		return nil, fmt.Errorf("failed to encrypt source config: %s", err)
	}

	// Use "source" category for source connectivity tests
	result, err := s.tempClient.TestConnection(ctx, "source", req.Type, req.Version, encryptedConfig)
	if err != nil {
		logs.Error("Connection test failed: %v", err)
	}

	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}

	return result, nil
}

func (s *SourceService) GetSourceCatalog(ctx context.Context, req dto.StreamsRequest) (map[string]interface{}, error) {
	if err := dto.Validate(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logs.Info("Getting source catalog for type=%s version=%s", req.Type, req.Version)
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available")
	}

	oldStreams := ""
	if req.JobID >= 0 {
		job, err := s.jobORM.GetByID(req.JobID, true)
		if err != nil {
			return nil, fmt.Errorf("job not found: %s", err)
		}
		oldStreams = job.StreamsConfig
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt config for source id: %d: %s", req.JobID, err)
	}
	// Use Temporal client to get the catalog
	var newStreams map[string]interface{}
	if s.tempClient != nil {
		newStreams, err = s.tempClient.GetCatalog(
			ctx,
			req.JobName,
			req.Type,
			req.Version,
			encryptedConfig,
			oldStreams,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog for source id: %d: %s", req.JobID, err)
	}

	return newStreams, nil
}

func (s *SourceService) GetSourceJobs(ctx context.Context, id int) ([]*models.Job, error) {
	if _, err := s.sourceORM.GetByID(id); err != nil {
		return nil, constants.ErrSourceNotFound
	}

	jobs, err := s.jobORM.GetBySourceID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source id: %d: %s", id, err)
	}

	return jobs, nil
}

func (s *SourceService) GetSourceVersions(ctx context.Context, sourceType string) (map[string]interface{}, error) {
	logs.Info("Getting source versions with source type: %s", sourceType)
	if sourceType == "" {
		return nil, fmt.Errorf("source type is required")
	}

	imageName := fmt.Sprintf("olakego/source-%s", sourceType)

	versions, _, err := utils.GetDriverImageTags(ctx, imageName, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker versions for source type %s: %s", sourceType, err)
	}

	return versions, nil
}

func (s *SourceService) GetSourceSpec(ctx context.Context, req dto.SpecRequest) (dto.SpecOutput, error) {
	logs.Info("Getting source spec for type: %s and version: %s", req.Type, req.Version)
	if req.Type == "" {
		return dto.SpecOutput{}, fmt.Errorf("source type is required")
	}
	if req.Version == "" {
		return dto.SpecOutput{}, fmt.Errorf("source version is required")
	}

	specOutput, err := s.tempClient.FetchSpec(ctx, "", req.Type, req.Version)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to fetch source spec: %s", err)
	}
	return specOutput, nil
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
