package etl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
)

// Source-related methods on AppService

// GetSource returns a single source by ID with its associated jobs.
func (s Service) GetSource(ctx context.Context, projectID string, sourceID int) (*dto.SourceDataItem, error) {
	source, err := s.db.GetSourceByID(sourceID)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrSourceNotFound, err)
		}
		return nil, fmt.Errorf("failed to get source: %s", err)
	}

	// Get jobs for this source
	jobs, err := s.db.GetJobsBySourceID([]int{sourceID})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for source: %s", err)
	}

	// Batch fetch workflow info for all jobs
	lastRunByJobID, err := fetchLatestJobRunsByJobIDs(ctx, s.temporal, projectID, jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest job runs from temporal: %s", err)
	}

	// Build job data items
	jobItems, err := buildJobDataItems(jobs, lastRunByJobID, "source")
	if err != nil {
		return nil, fmt.Errorf("failed to build job data items: %s", err)
	}

	item := &dto.SourceDataItem{
		ID:        source.ID,
		Name:      source.Name,
		Type:      source.Type,
		Version:   source.Version,
		Config:    source.Config,
		CreatedAt: source.CreatedAt.Format(time.RFC3339),
		UpdatedAt: source.UpdatedAt.Format(time.RFC3339),
		Jobs:      jobItems,
	}
	setUsernames(&item.CreatedBy, &item.UpdatedBy, source.CreatedBy, source.UpdatedBy)

	return item, nil
}

// GetAllSources returns all sources for a project with lightweight job summaries.
func (s Service) ListSources(ctx context.Context, projectID string) ([]dto.SourceDataItem, error) {
	sources, err := s.db.ListSourcesByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %s", err)
	}

	sourceIDs := make([]int, 0, len(sources))
	for _, src := range sources {
		sourceIDs = append(sourceIDs, src.ID)
	}

	var allJobs []*models.Job
	allJobs, err = s.db.GetJobsBySourceID(sourceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %s", err)
	}

	jobsBySourceID := make(map[int][]*models.Job)
	for _, job := range allJobs {
		jobsBySourceID[job.SourceID] = append(jobsBySourceID[job.SourceID], job)
	}

	// Batch fetch workflow info for all jobs
	lastRunByJobID, err := fetchLatestJobRunsByJobIDs(ctx, s.temporal, projectID, allJobs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest job runs from temporal: %s", err)
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
		jobItems, err := buildJobDataItems(jobs, lastRunByJobID, "source")
		if err != nil {
			return nil, fmt.Errorf("failed to build job data items: %s", err)
		}
		item.Jobs = jobItems

		items = append(items, item)
	}

	return items, nil
}

func (s Service) CreateSource(ctx context.Context, req *dto.CreateSourceRequest, projectID string, userID *int) error {
	unique, err := s.db.IsSourceNameUniqueInProject(ctx, projectID, req.Name)
	if err != nil {
		return fmt.Errorf("failed to check source name uniqueness: %s", err)
	}
	if !unique {
		return fmt.Errorf("source name '%s' is not unique", req.Name)
	}

	src := &models.Source{
		Name:      req.Name,
		Type:      req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}

	user := &models.User{ID: *userID}
	src.CreatedByID = user.ID
	src.UpdatedByID = user.ID
	src.CreatedBy = user
	src.UpdatedBy = user

	if err := s.db.CreateSource(src); err != nil {
		return fmt.Errorf("failed to create source: %s", err)
	}

	telemetry.TrackSourceCreation(ctx, src)
	return nil
}

func (s Service) UpdateSource(ctx context.Context, projectID string, id int, req *dto.UpdateSourceRequest, userID *int) error {
	existing, err := s.db.GetSourceByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			return fmt.Errorf("%w: %v", constants.ErrSourceNotFound, err)
		}
		return fmt.Errorf("failed to get source: %s", err)
	}

	existing.Name = req.Name
	existing.Config = req.Config
	existing.Type = req.Type
	existing.Version = req.Version

	user := &models.User{ID: *userID}
	existing.UpdatedByID = user.ID
	existing.UpdatedBy = user

	jobs, err := s.db.GetJobsBySourceID([]int{existing.ID})
	if err != nil {
		return fmt.Errorf("failed to fetch jobs for source update: %s", err)
	}

	if err := cancelAllJobWorkflows(ctx, s.temporal, jobs, projectID); err != nil {
		return fmt.Errorf("failed to cancel workflows for source update: %s", err)
	}

	if err := s.db.UpdateSource(existing); err != nil {
		return fmt.Errorf("failed to update source: %s", err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return nil
}

func (s Service) DeleteSource(ctx context.Context, id int) (*dto.DeleteSourceResponse, error) {
	src, err := s.db.GetSourceByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrSourceNotFound, err)
		}
		return nil, fmt.Errorf("failed to find source: %s", err)
	}

	jobs, err := s.db.GetJobsBySourceID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs for source deletion: %s", err)
	}
	if len(jobs) > 0 {
		return nil, fmt.Errorf("cannot delete source '%s' id[%d] because it is used in %d jobs; please delete the associated jobs first", src.Name, id, len(jobs))
	}

	if err := s.db.DeleteSource(id); err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrSourceNotFound, err)
		}
		return nil, fmt.Errorf("failed to delete source: %s", err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return &dto.DeleteSourceResponse{Name: src.Name}, nil
}

// TestSourceConnection starts a connection check and returns its operation ID. The check
// runs a connector container, so the caller polls the operation rather than holding the
// request open for it.
func (s Service) TestSourceConnection(ctx context.Context, projectID string, req *dto.SourceTestConnectionRequest) (string, error) {
	if s.temporal == nil {
		return "", fmt.Errorf("temporal client not available")
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt config for test connection: %s", err)
	}

	return s.temporal.StartVerifyDriverCredentials(ctx, projectID, "config", req.Type, req.Version, encryptedConfig)
}

// GetSourceCatalog starts catalog discovery and returns its operation ID.
func (s Service) GetSourceCatalog(ctx context.Context, projectID string, req *dto.StreamsRequest) (string, error) {
	oldStreams := ""
	if req.JobID >= 0 {
		job, err := s.db.GetJobByID(req.JobID, true)
		if err != nil {
			return "", fmt.Errorf("failed to find job for catalog: %s", err)
		}
		oldStreams = job.StreamsConfig
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt config for catalog: %s", err)
	}

	return s.temporal.StartDiscoverStreams(
		ctx,
		projectID,
		req.Type,
		req.Version,
		encryptedConfig,
		oldStreams,
		req.JobName,
		req.MaxDiscoverThreads,
	)
}

func (s Service) GetSourceVersions(ctx context.Context, sourceType string) (dto.VersionsResponse, error) {
	imageName := fmt.Sprintf("olakego/source-%s", sourceType)
	versions, _, err := utils.GetDriverImageTags(ctx, imageName, true)
	if err != nil {
		return dto.VersionsResponse{}, fmt.Errorf("failed to get Docker versions: %s", err)
	}

	return dto.VersionsResponse{Version: versions}, nil
}

// TODO: cache spec in db for each version
// GetSourceSpec starts a spec fetch and returns its operation ID.
func (s Service) GetSourceSpec(ctx context.Context, projectID string, req *dto.SpecRequest) (string, error) {
	return s.temporal.StartDriverSpecs(ctx, projectID, "", req.Type, req.Version, req.Type, req.Version)
}
