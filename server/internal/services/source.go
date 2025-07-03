package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
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
	tempClient, err := temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %s", err)
		return nil, fmt.Errorf("%s temporal client: %s", ErrFailedToCreate, err)
	}

	return &SourceService{
		sourceORM:  database.NewSourceORM(),
		userORM:    database.NewUserORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *SourceService) GetAllSources(projectID string) ([]models.SourceDataItem, error) {
	// Get all sources
	sources, err := s.sourceORM.GetAll()
	if err != nil {
		return nil, fmt.Errorf("%s sources: %s", ErrFailedToRetrieve, err)
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
			return nil, fmt.Errorf(ErrFormatFailedToGetJobs, err)
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
			return nil, fmt.Errorf("%s job data items: %s", ErrFailedToProcess, err)
		}
		item.Jobs = jobItems

		sourceItems = append(sourceItems, item)
	}

	return sourceItems, nil
}

func (s *SourceService) CreateSource(req models.CreateSourceRequest, projectID string, userID *int) error {
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
			return fmt.Errorf("failed to get user: %s", err)
		}
		source.CreatedBy = user
		source.UpdatedBy = user
	}

	if err := s.sourceORM.Create(source); err != nil {
		return fmt.Errorf("failed to create source: %s", err)
	}

	return nil
}

func (s *SourceService) UpdateSource(id int, req models.UpdateSourceRequest, userID *int) error {
	existingSource, err := s.sourceORM.GetByID(id)
	if err != nil {
		return fmt.Errorf("source not found: %s", err)
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
		return fmt.Errorf("failed to update source: %s", err)
	}

	return nil
}

func (s *SourceService) DeleteSource(id int) (*models.DeleteSourceResponse, error) {
	source, err := s.sourceORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("source not found: %s", err)
	}

	// Get all jobs using this source
	jobs, err := s.jobORM.GetBySourceID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for source: %s", err)
	}

	// Deactivate all jobs using this source
	for _, job := range jobs {
		job.Active = false
		if err := s.jobORM.Update(job); err != nil {
			return nil, fmt.Errorf("failed to deactivate job %d: %s", job.ID, err)
		}
	}

	// Delete the source
	if err := s.sourceORM.Delete(id); err != nil {
		return nil, fmt.Errorf("failed to delete source: %s", err)
	}

	return &models.DeleteSourceResponse{
		Name: source.Name,
	}, nil
}

func (s *SourceService) TestConnection(req models.SourceTestConnectionRequest) (map[string]interface{}, error) {
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available")
	}
	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt config: %s", err)
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

func (s *SourceService) GetSourceCatalog(req models.StreamsRequest) (map[string]interface{}, error) {
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

func (s *SourceService) GetSourceJobs(id int) ([]*models.Job, error) {
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

func (s *SourceService) GetSourceVersions(sourceType string) ([]string, error) {
	if sourceType == "" {
		return nil, fmt.Errorf("source type is required")
	}

	// Get versions from Docker Hub
	imageName := fmt.Sprintf("olakego/source-%s", sourceType)

	versions, err := utils.GetDockerHubTags(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker versions: %s", err)
	}

	return versions, nil
}

// Helper methods
func (s *SourceService) buildJobDataItems(jobs []*models.Job, _, _ string) ([]models.JobDataItem, error) {
	// This is a placeholder implementation - you'll need to implement this
	// based on your existing buildJobDataItems function logic
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

// You'll need to implement this helper function or move it to a utils package
func setUsernames(createdBy, updatedBy *string, createdByUser, updatedByUser *models.User) {
	// Implementation depends on your User model structure
	// This is a placeholder
	if createdByUser != nil {
		*createdBy = createdByUser.Username // or whatever field you use
	}
	if updatedByUser != nil {
		*updatedBy = updatedByUser.Username // or whatever field you use
	}
}
