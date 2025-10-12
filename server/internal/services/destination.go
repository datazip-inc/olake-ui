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

type DestinationService struct {
	destORM    *database.DestinationORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func NewDestinationService() (*DestinationService, error) {
	logs.Info("Creating destination service")
	tempClient, err := temporal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("%s temporal client: %s", constants.ErrFailedToCreate, err)
	}
	return &DestinationService{
		destORM:    database.NewDestinationORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *DestinationService) GetAllDestinations(ctx context.Context, projectID string) ([]dto.DestinationDataItem, error) {
	logs.Info("Retrieving destinations for project with id: %s", projectID)
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

	destItems := make([]dto.DestinationDataItem, 0, len(destinations))
	for _, dest := range destinations {
		item := dto.DestinationDataItem{
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

func (s *DestinationService) CreateDestination(ctx context.Context, req dto.CreateDestinationRequest, projectID string, userID *int) error {
	logs.Info("Creating destination: %s", req.Name)
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

	telemetry.TrackDestinationCreation(ctx, destination)
	return nil
}

func (s *DestinationService) UpdateDestination(ctx context.Context, id int, projectID string, req dto.UpdateDestinationRequest, userID *int) error {
	logs.Info("Updating destination with id: %d", id)
	existingDest, err := s.destORM.GetByID(id)
	if err != nil {
		return fmt.Errorf("destination not found: %s", err)
	}

	existingDest.Name = req.Name
	existingDest.DestType = req.Type
	existingDest.Version = req.Version
	existingDest.Config = req.Config
	if userID != nil {
		user := &models.User{ID: *userID}
		existingDest.UpdatedBy = user
	}

	// Cancel workflows for jobs linked to this destination before persisting change
	jobs, err := s.jobORM.GetByDestinationID(existingDest.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs for destination %d: %s", existingDest.ID, err)
	}

	for _, job := range jobs {
		if err := cancelJobWorkflow(s.tempClient, job, projectID); err != nil {
			return fmt.Errorf("failed to cancel workflow for job %d: %s", job.ID, err)
		}
	}

	if err := s.destORM.Update(existingDest); err != nil {
		return fmt.Errorf("failed to update destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return nil
}

func (s *DestinationService) DeleteDestination(ctx context.Context, id int) (*dto.DeleteDestinationResponse, error) {
	logs.Info("Deleting destination with id: %d", id)
	dest, err := s.destORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %s", err)
	}

	jobs, err := s.jobORM.GetByDestinationID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for destination: %s", err)
	}

	for _, job := range jobs {
		job.Active = false
		if err := s.jobORM.Update(job); err != nil {
			return nil, fmt.Errorf("failed to deactivate job %d: %s", job.ID, err)
		}
	}

	if err := s.destORM.Delete(id); err != nil {
		return nil, fmt.Errorf("failed to delete destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return &dto.DeleteDestinationResponse{Name: dest.Name}, nil
}

func (s *DestinationService) TestConnection(ctx context.Context, req dto.DestinationTestConnectionRequest) (map[string]interface{}, error) {
	logs.Info("Testing connection with config: %v", req.Config)
	if s.tempClient == nil {
		return nil, fmt.Errorf("temporal client not available")
	}

	// Determine driver and available tags
	version := req.Version
	driver := req.SourceType
	if driver == "" {
		var err error
		_, driver, err = utils.GetDriverImageTags(ctx, "", true)
		if err != nil {
			return nil, fmt.Errorf("failed to get valid driver image tags: %s", err)
		}
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("%s encrypt config: %s", constants.ErrFailedToProcess, err)
	}
	result, err := s.tempClient.TestConnection(ctx, "destination", driver, version, encryptedConfig)
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

func (s *DestinationService) GetDestinationJobs(ctx context.Context, id int) ([]*models.Job, error) {
	logs.Info("Retrieving jobs for destination with id: %d", id)
	if _, err := s.destORM.GetByID(id); err != nil {
		return nil, fmt.Errorf("destination not found: %s", err)
	}
	jobs, err := s.jobORM.GetByDestinationID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs: %s", err)
	}
	return jobs, nil
}

func (s *DestinationService) GetDestinationVersions(ctx context.Context, destType string) (map[string]interface{}, error) {
	logs.Info("Retrieving versions for destination type: %s", destType)
	if destType == "" {
		return nil, fmt.Errorf("destination type is required")
	}
	// get available driver versions
	versions, _, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch driver versions: %s", err)
	}
	return map[string]interface{}{"version": versions}, nil
}

// Helper function
func (s *DestinationService) buildJobDataItems(jobs []*models.Job, _, _ string) ([]dto.JobDataItem, error) {
	jobItems := make([]dto.JobDataItem, 0, len(jobs))
	for _, job := range jobs {
		jobItems = append(jobItems, dto.JobDataItem{
			ID:   job.ID,
			Name: job.Name,
		})
	}
	return jobItems, nil
}

func (s *DestinationService) GetDestinationSpec(ctx context.Context, req dto.SpecRequest) (dto.SpecResponse, error) {
	logs.Info("Getting destination spec for type: %s and version: %s", req.Type, req.Version)

	destinationType := "iceberg"
	if req.Type == "s3" {
		destinationType = "parquet"
	}

	// Determine driver and available tags
	_, driver, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get valid driver image tags: %s", err)
	}

	if s.tempClient == nil {
		return dto.SpecResponse{}, fmt.Errorf("temporal client not available")
	}

	specOutput, err := s.tempClient.FetchSpec(ctx, destinationType, driver, req.Version)
	if err != nil {
		logs.Error("Failed to get destination spec: %v", err)
		return dto.SpecResponse{}, fmt.Errorf("failed to get destination spec: %s", err)
	}

	return dto.SpecResponse{
		Type:    req.Type,
		Version: req.Version,
		Spec:    specOutput.Spec,
	}, nil
}
