package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
)

// Destination-related methods on AppService

// GetDestination returns a single destination by ID with its associated jobs.
func (s *ETLService) GetDestination(ctx context.Context, projectID string, destinationID int) (*dto.DestinationDataItem, error) {
	destination, err := s.db.GetDestinationByID(destinationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination: %s", err)
	}

	// Get jobs for this destination
	jobs, err := s.db.GetJobsByDestinationID([]int{destinationID})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for destination: %s", err)
	}

	// Batch fetch workflow info for all jobs
	lastRunByJobID, err := fetchLatestJobRunsByJobIDs(ctx, s.temporal, projectID, jobs)
	if err != nil {
		logger.Errorf("failed to fetch latest job runs from temporal project_id[%s] destination_id[%d]: %s", projectID, destinationID, err)
		lastRunByJobID = map[int]JobLastRunInfo{}
	}

	// Build job data items
	jobItems, err := buildJobDataItems(jobs, lastRunByJobID, "destination")
	if err != nil {
		return nil, fmt.Errorf("failed to build job data items: %s", err)
	}

	item := &dto.DestinationDataItem{
		ID:        destination.ID,
		Name:      destination.Name,
		Type:      destination.DestType,
		Version:   destination.Version,
		Config:    destination.Config,
		CreatedAt: destination.CreatedAt.Format(time.RFC3339),
		UpdatedAt: destination.UpdatedAt.Format(time.RFC3339),
		Jobs:      jobItems,
	}
	setUsernames(&item.CreatedBy, &item.UpdatedBy, destination.CreatedBy, destination.UpdatedBy)

	return item, nil
}

// ListDestinations returns all destinations for a project with lightweight job summaries.
func (s *ETLService) ListDestinations(ctx context.Context, projectID string) ([]dto.DestinationDataItem, error) {
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

	// Batch fetch workflow info for all jobs
	lastRunByJobID, err := fetchLatestJobRunsByJobIDs(ctx, s.temporal, projectID, allJobs)
	if err != nil {
		logger.Errorf("failed to fetch latest job runs from temporal project_id[%s]: %s", projectID, err)
		lastRunByJobID = map[int]JobLastRunInfo{}
	}

	destItems := make([]dto.DestinationDataItem, 0, len(destinations))
	for _, dest := range destinations {
		entity := dto.DestinationDataItem{
			ID:        dest.ID,
			Name:      dest.Name,
			Type:      dest.DestType,
			Version:   dest.Version,
			CreatedAt: dest.CreatedAt.Format(time.RFC3339),
			UpdatedAt: dest.UpdatedAt.Format(time.RFC3339),
		}
		setUsernames(&entity.CreatedBy, &entity.UpdatedBy, dest.CreatedBy, dest.UpdatedBy)

		jobs := jobsByDestID[dest.ID]
		jobItems, err := buildJobDataItems(jobs, lastRunByJobID, "destination")
		if err != nil {
			return nil, fmt.Errorf("failed to build job data items: %s", err)
		}
		entity.Jobs = jobItems
		destItems = append(destItems, entity)
	}

	return destItems, nil
}

func (s *ETLService) CreateDestination(ctx context.Context, req *dto.CreateDestinationRequest, projectID string, userID *int) error {
	unique, err := s.db.IsDestinationNameUniqueInProject(ctx, projectID, req.Name)
	if err != nil {
		return fmt.Errorf("failed to check destination name uniqueness: %s", err)
	}
	if !unique {
		return fmt.Errorf("destination name '%s' is not unique", req.Name)
	}

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

func (s *ETLService) UpdateDestination(ctx context.Context, id int, projectID string, req *dto.UpdateDestinationRequest, userID *int) error {
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

	if err := cancelAllJobWorkflows(ctx, s.temporal, jobs, projectID); err != nil {
		return fmt.Errorf("failed to cancel workflows for destination update: %s", err)
	}

	if err := s.db.UpdateDestination(existingDest); err != nil {
		return fmt.Errorf("failed to update destination: %s", err)
	}

	telemetry.TrackDestinationsStatus(ctx)
	return nil
}

func (s *ETLService) DeleteDestination(ctx context.Context, id int) (*dto.DeleteDestinationResponse, error) {
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

func (s *ETLService) TestDestinationConnection(ctx context.Context, req *dto.DestinationTestConnectionRequest) (map[string]interface{}, []map[string]interface{}, error) {
	version := req.Version
	driver := req.SourceType
	if driver == "" {
		var err error
		_, driver, err = utils.GetDriverImageTags(ctx, "", true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get driver image tags: %s", err)
		}
	}

	// Determine which config to use
	config := req.Config

	// If DestinationID is provided, fetch config from database
	if req.DestinationID > 0 {
		destination, err := s.db.GetDestinationByID(req.DestinationID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get destination config destination_id[%d]: %s", req.DestinationID, err)
		}
		config = destination.Config
		logger.Debugf("Using config from destination_id[%d] for test connection", req.DestinationID)
	} else if config == "" {
		return nil, nil, fmt.Errorf("either destination_id or config must be provided")
	}

	encryptedConfig, err := utils.Encrypt(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt config for test connection: %s", err)
	}
	workflowID := fmt.Sprintf("test-connection-%s-%d", req.Type, time.Now().Unix())
	result, err := s.temporal.VerifyDriverCredentials(ctx, workflowID, "destination", driver, version, encryptedConfig)
	// TODO: handle from frontend
	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}

	if err != nil {
		return result, nil, fmt.Errorf("connection test failed: %s", err)
	}

	homeDir := constants.DefaultConfigDir
	mainLogDir := filepath.Join(homeDir, workflowID)
	// Fetch the latest batch of logs by tailing from the end with default limit in the "older" direction.
	logs, err := utils.ReadLogs(mainLogDir, -1, -1, "older")
	if err != nil {
		return result, nil, fmt.Errorf("failed to read logs destination_type[%s] destination_version[%s] error[%s]",
			req.Type, req.Version, err)
	}

	return result, logs.Logs, nil
}

func (s *ETLService) GetDestinationJobs(_ context.Context, id int) ([]*models.Job, error) {
	if _, err := s.db.GetDestinationByID(id); err != nil {
		return nil, fmt.Errorf("failed to find destination: %s", err)
	}

	jobs, err := s.db.GetJobsByDestinationID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by destination: %s", err)
	}

	return jobs, nil
}

func (s *ETLService) GetDestinationVersions(ctx context.Context, destType string) (map[string]interface{}, error) {
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
func (s *ETLService) GetDestinationSpec(ctx context.Context, req *dto.SpecRequest) (dto.SpecResponse, error) {
	_, driver, err := utils.GetDriverImageTags(ctx, "", true)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get driver image tags: %s", err)
	}

	specOut, err := s.temporal.GetDriverSpecs(ctx, req.Type, driver, req.Version)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get spec: %s", err)
	}

	return dto.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOut.Spec,
	}, nil
}
