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

type DestinationService struct {
	destORM    *database.DestinationORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func NewDestinationService() (*DestinationService, error) {
	tempClient, err := temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
		return nil, fmt.Errorf("%s temporal client: %s", ErrFailedToCreate, err)
	}

	return &DestinationService{
		destORM:    database.NewDestinationORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *DestinationService) GetAllDestinations(projectID string) ([]models.DestinationDataItem, error) {
	// Get all destinations
	destinations, err := s.destORM.GetAllByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("%s destinations: %s", ErrFailedToRetrieve, err)
	}

	// Collect all destination IDs
	destIDs := make([]int, 0, len(destinations))
	for _, dest := range destinations {
		destIDs = append(destIDs, dest.ID)
	}

	// Get all jobs for all destinations in a single query
	var allJobs []*models.Job
	if len(destIDs) > 0 {
		allJobs, err = s.jobORM.GetByDestinationIDs(destIDs)
		if err != nil {
			return nil, fmt.Errorf(ErrFormatFailedToGetJobs, err)
		}
	}

	// Create a map of destination ID to jobs
	jobsByDestID := make(map[int][]*models.Job)
	for _, job := range allJobs {
		jobsByDestID[job.DestID.ID] = append(jobsByDestID[job.DestID.ID], job)
	}

	// Build the response
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

		// Get jobs for this destination from the pre-fetched map
		jobs := jobsByDestID[dest.ID]
		jobItems, err := s.buildJobDataItems(jobs, projectID, "destination")
		if err != nil {
			return nil, fmt.Errorf("%s job data items: %s", ErrFailedToProcess, err)
		}
		item.Jobs = jobItems

		destItems = append(destItems, item)
	}

	return destItems, nil
}

func (s *DestinationService) CreateDestination(req models.CreateDestinationRequest, projectID string, userID *int) error {
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
		return fmt.Errorf("failed to create destination: %s", err)
	}

	return nil
}

func (s *DestinationService) UpdateDestination(id int, req models.UpdateDestinationRequest, userID *int) error {
	existingDest, err := s.destORM.GetByID(id)
	if err != nil {
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
		return fmt.Errorf("failed to update destination: %s", err)
	}

	return nil
}

func (s *DestinationService) DeleteDestination(id int) (*models.DeleteDestinationResponse, error) {
	dest, err := s.destORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %s", err)
	}

	jobs, err := s.jobORM.GetByDestinationID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for destination: %s", err)
	}

	// Deactivate all jobs using this destination
	for _, job := range jobs {
		job.Active = false
		if err := s.jobORM.Update(job); err != nil {
			return nil, fmt.Errorf("failed to deactivate job %d: %s", job.ID, err)
		}
	}

	if err := s.destORM.Delete(id); err != nil {
		return nil, fmt.Errorf("failed to delete destination: %s", err)
	}

	return &models.DeleteDestinationResponse{
		Name: dest.Name,
	}, nil
}

func (s *DestinationService) TestConnection(req models.DestinationTestConnectionRequest) (map[string]interface{}, error) {
	if req.Type == "" {
		return nil, fmt.Errorf("destination type is required")
	}

	if req.Version == "" {
		return nil, fmt.Errorf("destination version is required")
	}
	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt destination config: %s", err)
	}
	result, err := s.tempClient.TestConnection(context.Background(), "destination", "postgres", "latest", encryptedConfig)
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

func (s *DestinationService) GetDestinationJobs(id int) ([]*models.Job, error) {
	// Check if destination exists
	_, err := s.destORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %s", err)
	}

	jobs, err := s.jobORM.GetByDestinationID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs: %s", err)
	}

	return jobs, nil
}

func (s *DestinationService) GetDestinationVersions(destType string) ([]string, error) {
	if destType == "" {
		return nil, fmt.Errorf("destination type is required")
	}

	// In a real implementation, we would query for available versions
	// based on the destination type and project ID
	// For now, we'll return a mock response
	versions := []string{"latest"}

	return versions, nil
}

// Helper methods
func (s *DestinationService) buildJobDataItems(jobs []*models.Job, _, _ string) ([]models.JobDataItem, error) {
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
