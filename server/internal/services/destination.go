package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/models/dto"
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
	tempClient, err := temporal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client - error=%v", err)
	}
	return &DestinationService{
		destORM:    database.NewDestinationORM(),
		jobORM:     database.NewJobORM(),
		tempClient: tempClient,
	}, nil
}

func (s *DestinationService) GetAllDestinations(_ context.Context, projectID string) ([]dto.DestinationDataItem, error) {
	destinations, err := s.destORM.GetAllByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve destinations - project_id=%s error=%v", projectID, err)
	}

	destIDs := make([]int, 0, len(destinations))
	for _, dest := range destinations {
		destIDs = append(destIDs, dest.ID)
	}

	var allJobs []*models.Job
	allJobs, err = s.jobORM.GetByDestinationID(destIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs - project_id=%s destination_count=%d error=%v",
			projectID, len(destIDs), err)
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
		jobItems, err := s.buildJobDataItems(jobs, projectID, "destination")
		if err != nil {
			return nil, fmt.Errorf("failed to process job data items - project_id=%s destination_id=%d error=%v",
				projectID, dest.ID, err)
		}
		entity.Jobs = jobItems
		destItems = append(destItems, entity)
	}

	return destItems, nil
}

func (s *DestinationService) CreateDestination(ctx context.Context, req *dto.CreateDestinationRequest, projectID string, userID *int) error {
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

	if err := s.destORM.Create(destination); err != nil {
		return fmt.Errorf("failed to create destination - project_id=%s destination_name=%s destination_type=%s user_id=%v error=%v",
			projectID, req.Name, req.Type, userID, err)
	}

	telemetry.TrackDestinationCreation(ctx, destination)
	return nil
}

func (s *DestinationService) UpdateDestination(ctx context.Context, id int, projectID string, req *dto.UpdateDestinationRequest, userID *int) error {
	existingDest, err := s.destORM.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to find destination for update - project_id=%s destination_id=%d error=%v",
			projectID, id, err)
	}

	existingDest.Name = req.Name
	existingDest.DestType = req.Type
	existingDest.Version = req.Version
	existingDest.Config = req.Config

	user := &models.User{ID: *userID}
	existingDest.UpdatedBy = user

	jobs, err := s.jobORM.GetByDestinationID([]int{existingDest.ID})
	if err != nil {
		return fmt.Errorf("failed to fetch jobs for destination update - project_id=%s destination_id=%d error=%v",
			projectID, existingDest.ID, err)
	}

	if err := cancelAllJobWorkflows(s.tempClient, jobs, projectID); err != nil {
		return fmt.Errorf("failed to cancel workflows for destination update - project_id=%s destination_id=%d job_count=%d error=%v",
			projectID, existingDest.ID, len(jobs), err)
	}

	if err := s.destORM.Update(existingDest); err != nil {
		return fmt.Errorf("failed to update destination - project_id=%s destination_id=%d destination_name=%s error=%v",
			projectID, id, req.Name, err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return nil
}

func (s *DestinationService) DeleteDestination(ctx context.Context, id int) (*dto.DeleteDestinationResponse, error) {
	dest, err := s.destORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find destination for deletion - destination_id=%d error=%v", id, err)
	}

	jobs, err := s.jobORM.GetByDestinationID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs for destination deletion - destination_id=%d error=%v", id, err)
	}
	if len(jobs) > 0 {
		return nil, fmt.Errorf("cannot delete destination '%s' (id=%d) because it is used in %d jobs. Please delete the associated jobs first.", dest.Name, id, len(jobs))
	}
	var jobIDs []int
	for _, job := range jobs {
		job.Active = false
		jobIDs = append(jobIDs, job.ID)
	}

	if err := s.jobORM.UpdateAllJobs(jobIDs); err != nil {
		return nil, fmt.Errorf("failed to deactivate jobs for destination deletion - destination_id=%d job_count=%d error=%v",
			id, len(jobIDs), err)
	}

	if err := s.destORM.Delete(id); err != nil {
		return nil, fmt.Errorf("failed to delete destination - destination_id=%d destination_name=%s error=%v",
			id, dest.Name, err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return &dto.DeleteDestinationResponse{Name: dest.Name}, nil
}

func (s *DestinationService) TestConnection(ctx context.Context, req *dto.DestinationTestConnectionRequest) (map[string]interface{}, []map[string]interface{}, error) {
	version := req.Version
	driver := req.SourceType
	if driver == "" {
		var err error
		_, driver, err = utils.GetDriverImageTags(ctx, "", true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get driver image tags - destination_type=%s error=%v", req.Type, err)
		}
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt config for test connection - destination_type=%s destination_version=%s error=%v",
			req.Type, req.Version, err)
	}
	workflowID := fmt.Sprintf("test-connection-%s-%d", req.Type, time.Now().Unix())
	result, err := s.tempClient.TestConnection(ctx, workflowID, "destination", driver, version, encryptedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connection test failed - destination_type=%s destination_version=%s error=%v",
			req.Type, req.Version, err)
	}
	homeDir := docker.GetDefaultConfigDir()
	mainLogDir := filepath.Join(homeDir, workflowID)
	logs, err := utils.ReadLogs(mainLogDir)
	if err != nil {
		logger.Error("failed to read logs - destination_type=%s destination_version=%s error=%v",
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

func (s *DestinationService) GetDestinationJobs(_ context.Context, id int) ([]*models.Job, error) {
	if _, err := s.destORM.GetByID(id); err != nil {
		return nil, fmt.Errorf("failed to find destination - destination_id=%d error=%v", id, err)
	}

	jobs, err := s.jobORM.GetByDestinationID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by destination - destination_id=%d error=%v", id, err)
	}

	return jobs, nil
}

func (s *DestinationService) GetDestinationVersions(ctx context.Context, destType string) (map[string]interface{}, error) {
	if destType == "" {
		return nil, fmt.Errorf("destination type is required")
	}

	versions, _, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver image tags - destination_type=%s error=%v", destType, err)
	}

	return map[string]interface{}{"version": versions}, nil
}

// TODO: cache spec in db for each version
func (s *DestinationService) GetDestinationSpec(ctx context.Context, req *dto.SpecRequest) (dto.SpecResponse, error) {
	// TODO: handle from frontend
	destinationType := "iceberg"
	if req.Type == "s3" {
		destinationType = "parquet"
	}

	_, driver, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get driver image tags - destination_type=%s error=%v", req.Type, err)
	}

	specOut, err := s.tempClient.FetchSpec(ctx, destinationType, driver, req.Version)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get spec - destination_type=%s destination_version=%s error=%v",
			req.Type, req.Version, err)
	}

	return dto.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOut.Spec,
	}, nil
}

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
