package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
)

// Destination-related methods on AppService

// ListDestinations returns all destinations for a project with lightweight job summaries.
func (s *AppService) ListDestinations(_ context.Context, projectID string) ([]dto.DestinationDataItem, error) {
	destinations, err := s.db.ListDestinationsByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list destinations: %s", err)
	}

	destIDs := make([]int, 0, len(destinations))
	for _, dest := range destinations {
		destIDs = append(destIDs, dest.ID)
	}

	var allJobs []*models.Job
	allJobs, err = s.db.GetJobsByDestinationID(destIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %s", err)
	}

	jobsByDestID := make(map[int][]*models.Job)
	for _, job := range allJobs {
		jobsByDestID[job.DestID.ID] = append(jobsByDestID[job.DestID.ID], job)
	}

	destItems := make([]dto.DestinationDataItem, 0, len(destinations))
	for _, dest := range destinations {
		entity := dto.DestinationDataItem{
			ID:        dest.ID,
			Name:      dest.Name,
			Type:      dest.DestType,
			Version:   dest.Version,
			Config:    dest.Config,
			CreatedAt: dest.CreatedAt.Format(time.RFC3339),
			UpdatedAt: dest.UpdatedAt.Format(time.RFC3339),
		}
		setUsernames(&entity.CreatedBy, &entity.UpdatedBy, dest.CreatedBy, dest.UpdatedBy)

		jobs := jobsByDestID[dest.ID]
		jobItems, err := buildJobDataItems(jobs)
		if err != nil {
			return nil, fmt.Errorf("failed to build job data items: %s", err)
		}
		entity.Jobs = jobItems
		destItems = append(destItems, entity)
	}

	return destItems, nil
}

func (s *AppService) CreateDestination(ctx context.Context, req *dto.CreateDestinationRequest, projectID string, userID *int) error {
	destination := &models.Destination{
		Name:      req.Name,
		DestType:  req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}
	user := &models.User{ID: *userID}
	destination.CreatedBy = user
	destination.UpdatedBy = user

	if err := s.db.CreateDestination(destination); err != nil {
		return fmt.Errorf("failed to create destination: %s", err)
	}

	telemetry.TrackDestinationCreation(ctx, destination)
	return nil
}

func (s *AppService) UpdateDestination(ctx context.Context, id int, projectID string, req *dto.UpdateDestinationRequest, userID *int) error {
	existingDest, err := s.db.GetDestinationByID(id)
	if err != nil {
		return fmt.Errorf("failed to get destination: %s", err)
	}

	existingDest.Name = req.Name
	existingDest.DestType = req.Type
	existingDest.Version = req.Version
	existingDest.Config = req.Config

	user := &models.User{ID: *userID}
	existingDest.UpdatedBy = user

	jobs, err := s.db.GetJobsByDestinationID([]int{existingDest.ID})
	if err != nil {
		return fmt.Errorf("failed to fetch jobs for destination update: %s", err)
	}

	if err := cancelAllJobWorkflows(ctx, s.tempClient, jobs, projectID); err != nil {
		return fmt.Errorf("failed to cancel workflows for destination update: %s", err)
	}

	if err := s.db.UpdateDestination(existingDest); err != nil {
		return fmt.Errorf("failed to update destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return nil
}

func (s *AppService) DeleteDestination(ctx context.Context, id int) (*dto.DeleteDestinationResponse, error) {
	dest, err := s.db.GetDestinationByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find destination: %s", err)
	}

	jobs, err := s.db.GetJobsByDestinationID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs for destination deletion: %s", err)
	}
	if len(jobs) > 0 {
		return nil, fmt.Errorf("cannot delete destination '%s' id[%d] because it is used in %d jobs; please delete the associated jobs first", dest.Name, id, len(jobs))
	}
	var jobIDs []int
	for _, job := range jobs {
		job.Active = false
		jobIDs = append(jobIDs, job.ID)
	}

	if err := s.db.DeactivateJobs(jobIDs); err != nil {
		return nil, fmt.Errorf("failed to deactivate jobs for destination deletion: %s", err)
	}

	if err := s.db.DeleteDestination(id); err != nil {
		return nil, fmt.Errorf("failed to delete destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return &dto.DeleteDestinationResponse{Name: dest.Name}, nil
}

func (s *AppService) TestConnection(ctx context.Context, req *dto.DestinationTestConnectionRequest) (map[string]interface{}, []map[string]interface{}, error) {
	version := req.Version
	driver := req.SourceType
	if driver == "" {
		var err error
		_, driver, err = utils.GetDriverImageTags(ctx, "", true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get driver image tags: %s", err)
		}
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt config for test connection: %s", err)
	}
	workflowID := fmt.Sprintf("test-connection-%s-%d", req.Type, time.Now().Unix())
	result, err := s.tempClient.TestConnection(ctx, workflowID, "destination", driver, version, encryptedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connection test failed: %s", err)
	}
	homeDir := docker.GetDefaultConfigDir()
	mainLogDir := filepath.Join(homeDir, workflowID)
	logs, err := utils.ReadLogs(mainLogDir)
	if err != nil {
		logger.Error("failed to read logs destination_type[%s] destination_version[%s] error[%s]",
			req.Type, req.Version, err)
	}
	// TODO: handle from frontend
	if result == nil {
		result = map[string]interface{}{
			"message": "Connection test failed",
			"status":  "failed",
		}
	}

	return result, logs, nil
}

func (s *AppService) GetDestinationJobs(_ context.Context, id int) ([]*models.Job, error) {
	if _, err := s.db.GetDestinationByID(id); err != nil {
		return nil, fmt.Errorf("failed to find destination: %s", err)
	}

	jobs, err := s.db.GetJobsByDestinationID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by destination: %s", err)
	}

	return jobs, nil
}

func (s *AppService) GetDestinationVersions(ctx context.Context, destType string) (map[string]interface{}, error) {
	if destType == "" {
		return nil, fmt.Errorf("destination type is required")
	}

	versions, _, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver image tags: %s", err)
	}

	return map[string]interface{}{"version": versions}, nil
}

// TODO: cache spec in db for each version
func (s *AppService) GetDestinationSpec(ctx context.Context, req *dto.SpecRequest) (dto.SpecResponse, error) {
	// TODO: handle from frontend
	destinationType := "iceberg"
	if req.Type == "s3" {
		destinationType = "parquet"
	}

	_, driver, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get driver image tags: %s", err)
	}

	specOut, err := s.tempClient.FetchSpec(ctx, destinationType, driver, req.Version)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get spec: %s", err)
	}

	return dto.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOut.Spec,
	}, nil
}
