package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/internal/temporal"
	"github.com/datazip/olake-frontend/server/utils"
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

func (s *SourceService) GetAllSources(ctx context.Context, projectID string) ([]models.SourceDataItem, error) {
	logs.Info("Getting all sources")
	// Get all sources
	sources, err := s.sourceORM.GetAll()
	if err != nil {
		return nil, fmt.Errorf("%s sources: %s", constants.ErrFailedToRetrieve, err)
	}

	// Collect all source IDs
	sourceIDs := make([]int, 0, len(sources))
	for _, source := range sources {
		sourceIDs = append(sourceIDs, source.ID)
	}

	// Get all jobs for all sources in a single query
	var allJobs []*models.Job
	if len(sourceIDs) > 0 {
		allJobs, err = s.jobORM.GetBySourceIDs(sourceIDs)
		if err != nil {
			return nil, fmt.Errorf("%s jobs: %s", constants.ErrFailedToRetrieve, err)
		}
	}

	// Create a map of source ID to jobs
	jobsBySourceID := make(map[int][]*models.Job)
	for _, job := range allJobs {
		jobsBySourceID[job.SourceID.ID] = append(jobsBySourceID[job.SourceID.ID], job)
	}

	// Build the response
	sourceItems := make([]models.SourceDataItem, 0, len(sources))
	for _, source := range sources {
		item := models.SourceDataItem{
			ID:        source.ID,
			Name:      source.Name,
			Type:      source.Type,
			Version:   source.Version,
			Config:    source.Config,
			CreatedAt: source.CreatedAt.Format(time.RFC3339),
			UpdatedAt: source.UpdatedAt.Format(time.RFC3339),
		}

		setUsernames(&item.CreatedBy, &item.UpdatedBy, source.CreatedBy, source.UpdatedBy)

		// Get jobs for this source from the pre-fetched map
		jobs := jobsBySourceID[int(source.ID)]
		jobItems, err := s.buildJobDataItems(jobs, projectID, "source")
		if err != nil {
			return nil, fmt.Errorf("%s job data items: %s", constants.ErrFailedToProcess, err)
		}
		item.Jobs = jobItems

		sourceItems = append(sourceItems, item)
	}

	return sourceItems, nil
}

func (s *SourceService) CreateSource(ctx context.Context, req models.CreateSourceRequest, projectID string, userID *int) error {
	logs.Info("Creating source with projectID: %s", projectID)
	source := &models.Source{
		Name:      req.Name,
		Type:      req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}

	if userID != nil {
		user, err := s.userORM.GetByID(*userID)
		if err != nil {
			return fmt.Errorf("%s user: %s", constants.ErrFailedToRetrieve, err)
		}
		source.CreatedBy = user
		source.UpdatedBy = user
	}

	if err := s.sourceORM.Create(source); err != nil {
		return fmt.Errorf("%s source: %s", constants.ErrFailedToCreate, err)
	}
	telemetry.TrackSourceCreation(context.Background(), source)
	return nil
}

func (s *SourceService) UpdateSource(ctx context.Context, id int, req models.UpdateSourceRequest, userID *int) error {
	logs.Info("Updating source with id: %d", id)
	existingSource, err := s.sourceORM.GetByID(id)
	if err != nil {
		return fmt.Errorf("%s source: %s", constants.ErrFailedToRetrieve, err)
	}

	existingSource.Name = req.Name
	existingSource.Config = req.Config
	existingSource.Type = req.Type
	existingSource.Version = req.Version
	existingSource.UpdatedAt = time.Now()

	if userID != nil {
		user := &models.User{ID: *userID}
		existingSource.UpdatedBy = user
	}

	if err := s.sourceORM.Update(existingSource); err != nil {
		return fmt.Errorf("%s source: %s", constants.ErrFailedToUpdate, err)
	}
	telemetry.TrackSourcesStatus(ctx)
	return nil
}

func (s *SourceService) DeleteSource(ctx context.Context, id int) (*models.DeleteSourceResponse, error) {
	logs.Info("Deleting source with id: %d", id)
	source, err := s.sourceORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%s source: %s", constants.ErrFailedToRetrieve, err)
	}

	// Get all jobs using this source
	jobs, err := s.jobORM.GetBySourceID(id)
	if err != nil {
		return nil, fmt.Errorf("%s jobs for source: %s", constants.ErrFailedToRetrieve, err)
	}

	// Get job IDs to deactivate
	jobIDs := make([]int, 0, len(jobs))
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	// Deactivate all jobs using this source in a single query
	if len(jobIDs) > 0 {
		if err := s.jobORM.UpdateAllJobs(jobIDs); err != nil {
			return nil, fmt.Errorf("%s jobs: %s", constants.ErrFailedToUpdate, err)
		}
	}

	// Delete the source
	if err := s.sourceORM.Delete(id); err != nil {
		return nil, fmt.Errorf("%s source: %s", constants.ErrFailedToDelete, err)
	}
	telemetry.TrackSourcesStatus(ctx)
	return &models.DeleteSourceResponse{
		Name: source.Name,
	}, nil
}

func (s *SourceService) TestConnection(ctx context.Context, req models.SourceTestConnectionRequest) (map[string]interface{}, error) {
	//logs.Info("Testing connection with config: %v", req.Config)
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("%s encrypt config: %s", constants.ErrFailedToProcess, err)
	}
	result, err := s.tempClient.TestConnection(context.Background(), "config", "postgres", "latest", encryptedConfig)
	if err != nil {
		logs.Error("Connection test failed: %v", err)
	}

	if result == nil {
		result = map[string]interface{}{
			"message": "Connection test failed: Please check your configuration and try again",
			"status":  "failed",
		}
	}

	return result, nil
}

func (s *SourceService) GetSourceCatalog(ctx context.Context, req models.StreamsRequest) (map[string]interface{}, error) {
	//logs.Info("Getting source catalog with config: %v", req.Config)
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	oldStreams := ""
	// Load job details if JobID is provided
	if req.JobID >= 0 {
		job, err := s.jobORM.GetByID(req.JobID, true)
		if err != nil {
			return nil, fmt.Errorf("job not found: %s", err)
		}
		oldStreams = job.StreamsConfig
	}
	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt config: %s", err)
	}
	// Use Temporal client to get the catalog
	var newStreams map[string]interface{}
	if s.tempClient != nil {
		newStreams, err = s.tempClient.GetCatalog(
			context.Background(),
			req.Type,
			req.Version,
			encryptedConfig,
			oldStreams,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %s", err)
	}

	return newStreams, nil
}

func (s *SourceService) GetSourceJobs(ctx context.Context, id int) ([]*models.Job, error) {
	// Check if source exists
	_, err := s.sourceORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("source not found: %s", err)
	}

	jobs, err := s.jobORM.GetBySourceID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source ID: %s", err)
	}

	return jobs, nil
}

func (s *SourceService) GetSourceVersions(ctx context.Context, sourceType string) ([]string, error) {
	logs.Info("Getting source versions with source type: %s", sourceType)
	if sourceType == "" {
		return nil, fmt.Errorf("source type is required")
	}

	// Get versions from Docker Hub
	imageName := fmt.Sprintf("olakego/source-%s", sourceType)

	versions, err := utils.GetDockerHubTags(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker versions for source type %s: %s", sourceType, err)
	}

	return versions, nil
}

// Helper methods
func (s *SourceService) buildJobDataItems(jobs []*models.Job, _, _ string) ([]models.JobDataItem, error) {
	jobItems := make([]models.JobDataItem, 0, len(jobs))

	for _, job := range jobs {
		item := models.JobDataItem{
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
