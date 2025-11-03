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
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
)

// Source-related methods on AppService

// GetAllSources returns all sources for a project with lightweight job summaries.
func (s *ETLService) ListSources(_ context.Context, _ string) ([]dto.SourceDataItem, error) {
	sources, err := s.db.ListSources()
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
		jobItems, err := buildJobDataItems(jobs, s.temporal, "source")
		if err != nil {
			return nil, fmt.Errorf("failed to build job data items: %s", err)
		}
		item.Jobs = jobItems

		items = append(items, item)
	}

	return items, nil
}

func (s *ETLService) CreateSource(ctx context.Context, req *dto.CreateSourceRequest, projectID string, userID *int) error {
	src := &models.Source{
		Name:      req.Name,
		Type:      req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectID,
	}

	user := &models.User{ID: *userID}
	src.CreatedBy = user
	src.UpdatedBy = user

	if err := s.db.CreateSource(src); err != nil {
		return fmt.Errorf("failed to create source: %s", err)
	}

	telemetry.TrackSourceCreation(ctx, src)
	return nil
}

func (s *ETLService) UpdateSource(ctx context.Context, projectID string, id int, req *dto.UpdateSourceRequest, userID *int) error {
	existing, err := s.db.GetSourceByID(id)
	if err != nil {
		return fmt.Errorf("failed to get source: %s", err)
	}

	existing.Name = req.Name
	existing.Config = req.Config
	existing.Type = req.Type
	existing.Version = req.Version

	user := &models.User{ID: *userID}
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

func (s *ETLService) DeleteSource(ctx context.Context, id int) (*dto.DeleteSourceResponse, error) {
	src, err := s.db.GetSourceByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find source: %s", err)
	}

	jobs, err := s.db.GetJobsBySourceID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs for source deletion: %s", err)
	}
	if len(jobs) > 0 {
		return nil, fmt.Errorf("cannot delete source '%s' id[%d] because it is used in %d jobs; please delete the associated jobs first", src.Name, id, len(jobs))
	}
	jobIDs := make([]int, 0, len(jobs))
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	if err := s.db.DeactivateJobs(jobIDs); err != nil {
		return nil, fmt.Errorf("failed to update jobs for source deletion: %s", err)
	}

	if err := s.db.DeleteSource(id); err != nil {
		return nil, fmt.Errorf("failed to delete source: %s", err)
	}

	telemetry.TrackSourcesStatus(ctx)
	return &dto.DeleteSourceResponse{Name: src.Name}, nil
}

func (s *ETLService) TestSourceConnection(ctx context.Context, req *dto.SourceTestConnectionRequest) (map[string]interface{}, []map[string]interface{}, error) {
	if s.temporal == nil {
		return nil, nil, fmt.Errorf("temporal client not available")
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt config for test connection: %s", err)
	}
	workflowID := fmt.Sprintf("test-connection-%s-%d", req.Type, time.Now().Unix())
	result, err := s.temporal.VerifyDriverCredentials(ctx, workflowID, "config", req.Type, req.Version, encryptedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connection test failed: %s", err)
	}
	homeDir := constants.DefaultConfigDir
	mainLogDir := filepath.Join(homeDir, workflowID)
	logs, err := utils.ReadLogs(mainLogDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read logs source_type[%s] source_version[%s]: %s",
			req.Type, req.Version, err)
	}
	// TODO: handle from frontend
	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}

	return result, logs, nil
}

func (s *ETLService) GetSourceCatalog(ctx context.Context, req *dto.StreamsRequest) (map[string]interface{}, error) {
	oldStreams := ""
	if req.JobID >= 0 {
		job, err := s.db.GetJobByID(req.JobID, true)
		if err != nil {
			return nil, fmt.Errorf("failed to find job for catalog: %s", err)
		}
		oldStreams = job.StreamsConfig
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt config for catalog: %s", err)
	}

	newStreams, err := s.temporal.DiscoverStreams(
		ctx,
		req.Type,
		req.Version,
		encryptedConfig,
		oldStreams,
		req.JobName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %s", err)
	}

	return newStreams, nil
}

func (s *ETLService) GetSourceJobs(_ context.Context, id int) ([]*models.Job, error) {
	if _, err := s.db.GetSourceByID(id); err != nil {
		return nil, fmt.Errorf("failed to find source: %s", err)
	}

	jobs, err := s.db.GetJobsBySourceID([]int{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source: %s", err)
	}

	return jobs, nil
}

func (s *ETLService) GetSourceVersions(ctx context.Context, sourceType string) (map[string]interface{}, error) {
	if sourceType == "" {
		return nil, fmt.Errorf("source type is required")
	}

	imageName := fmt.Sprintf("olakego/source-%s", sourceType)
	versions, _, err := utils.GetDriverImageTags(ctx, imageName, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker versions: %s", err)
	}

	return map[string]interface{}{"version": versions}, nil
}

// TODO: cache spec in db for each version
func (s *ETLService) GetSourceSpec(ctx context.Context, req *dto.SpecRequest) (dto.SpecResponse, error) {
	specOut, err := s.temporal.GetDriverSpecs(ctx, "", req.Type, req.Version)
	if err != nil {
		return dto.SpecResponse{}, fmt.Errorf("failed to get spec: %s", err)
	}

	return dto.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOut.Spec,
	}, nil
}
