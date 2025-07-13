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

type DestinationService struct {
	destORM    *database.DestinationORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func NewDestinationService() (*DestinationService, error) {
	tempClient, err := temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
		return nil, fmt.Errorf("%s temporal client: %s", constants.ErrFailedToCreate, err)
	}

	return &DestinationService{
		destORM:    database.NewDestinationORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *DestinationService) GetAllDestinations(ctx context.Context, projectID string) ([]models.DestinationDataItem, error) {
	destinations, err := s.destORM.GetAllByProjectID(projectID)
	if err != nil {
		logs.Error("Failed to retrieve destinations: %v", err)
		return nil, fmt.Errorf("%s destinations: %s", constants.ErrFailedToRetrieve, err)
	}

	destIDs := make([]int, 0, len(destinations))
	for _, dest := range destinations {
		destIDs = append(destIDs, dest.ID)
	}

	var allJobs []*models.Job
	if len(destIDs) > 0 {
		allJobs, err = s.jobORM.GetByDestinationIDs(destIDs)
		if err != nil {
			logs.Error("Failed to get jobs: %v", err)
			return nil, fmt.Errorf(constants.ErrFormatFailedToGetJobs, err)
		}
	}

	jobsByDestID := make(map[int][]*models.Job)
	for _, job := range allJobs {
		jobsByDestID[job.DestID.ID] = append(jobsByDestID[job.DestID.ID], job)
	}

	destItems := make([]models.DestinationDataItem, 0, len(destinations))
	for _, dest := range destinations {
		item := models.DestinationDataItem{
			ID:        dest.ID,
			Name:      dest.Name,
			Type:      dest.DestType,
			Version:   dest.Version,
			Config:    dest.Config,
			CreatedAt: dest.CreatedAt.Format(time.RFC3339),
			UpdatedAt: dest.UpdatedAt.Format(time.RFC3339),
		}

		setUsernames(&item.CreatedBy, &item.UpdatedBy, dest.CreatedBy, dest.UpdatedBy)

		jobs := jobsByDestID[dest.ID]
		jobItems, err := s.buildJobDataItems(jobs, projectID, "destination")
		if err != nil {
			logs.Error("Failed to build job items: %v", err)
			return nil, fmt.Errorf("%s job data items: %s", constants.ErrFailedToProcess, err)
		}
		item.Jobs = jobItems

		destItems = append(destItems, item)
	}

	return destItems, nil
}

func (s *DestinationService) CreateDestination(ctx context.Context, req models.CreateDestinationRequest, projectID string, userID *int) error {
	destination := &models.Destination{
		Name:      req.Name,
		DestType:  req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}
	if userID != nil {
		user := &models.User{ID: *userID}
		destination.CreatedBy = user
		destination.UpdatedBy = user
	}

	if err := s.destORM.Create(destination); err != nil {
		logs.Error("Failed to create destination: %v", err)
		return fmt.Errorf("failed to create destination: %s", err)
	}

	telemetry.TrackDestinationCreation(ctx, destination)
	return nil
}

func (s *DestinationService) UpdateDestination(ctx context.Context, id int, req models.UpdateDestinationRequest, userID *int) error {
	existingDest, err := s.destORM.GetByID(id)
	if err != nil {
		logs.Warn("Destination not found: %v", err)
		return fmt.Errorf("destination not found: %s", err)
	}

	existingDest.Name = req.Name
	existingDest.DestType = req.Type
	existingDest.Version = req.Version
	existingDest.Config = req.Config
	existingDest.UpdatedAt = time.Now()
	if userID != nil {
		user := &models.User{ID: *userID}
		existingDest.UpdatedBy = user
	}

	if err := s.destORM.Update(existingDest); err != nil {
		logs.Error("Failed to update destination: %v", err)
		return fmt.Errorf("failed to update destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return nil
}

func (s *DestinationService) DeleteDestination(ctx context.Context, id int) (*models.DeleteDestinationResponse, error) {
	dest, err := s.destORM.GetByID(id)
	if err != nil {
		logs.Warn("Destination not found: %v", err)
		return nil, fmt.Errorf("destination not found: %s", err)
	}

	jobs, err := s.jobORM.GetByDestinationID(id)
	if err != nil {
		logs.Error("Failed to get jobs for destination: %v", err)
		return nil, fmt.Errorf("failed to get jobs for destination: %s", err)
	}

	for _, job := range jobs {
		job.Active = false
		if err := s.jobORM.Update(job); err != nil {
			logs.Error("Failed to deactivate job %d: %v", job.ID, err)
			return nil, fmt.Errorf("failed to deactivate job %d: %s", job.ID, err)
		}
	}

	if err := s.destORM.Delete(id); err != nil {
		logs.Error("Failed to delete destination: %v", err)
		return nil, fmt.Errorf("failed to delete destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return &models.DeleteDestinationResponse{Name: dest.Name}, nil
}

func (s *DestinationService) TestConnection(ctx context.Context, req models.DestinationTestConnectionRequest) (map[string]interface{}, error) {
	if req.Type == "" || req.Version == "" {
		return nil, fmt.Errorf("destination type and version are required")
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		logs.Error("Failed to encrypt config: %v", err)
		return nil, fmt.Errorf("failed to encrypt destination config: %s", err)
	}

	result, err := s.tempClient.TestConnection(ctx, "destination", req.Type, req.Version, encryptedConfig)
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

func (s *DestinationService) GetDestinationJobs(ctx context.Context, id int) ([]*models.Job, error) {
	_, err := s.destORM.GetByID(id)
	if err != nil {
		logs.Warn("Destination not found: %v", err)
		return nil, fmt.Errorf("destination not found: %s", err)
	}

	jobs, err := s.jobORM.GetByDestinationID(id)
	if err != nil {
		logs.Error("Failed to retrieve jobs: %v", err)
		return nil, fmt.Errorf("failed to retrieve jobs: %s", err)
	}

	return jobs, nil
}

func (s *DestinationService) GetDestinationVersions(destType string) ([]string, error) {
	if destType == "" {
		return nil, fmt.Errorf("destination type is required")
	}
	return []string{"latest"}, nil
}

// Helper function
func (s *DestinationService) buildJobDataItems(jobs []*models.Job, _, _ string) ([]models.JobDataItem, error) {
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
