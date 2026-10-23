package etl

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
)

// ConvertLegacyJobs migrates legacy (Streams V1) jobs to the split format, each with its source's
// version. A job that fails stays legacy; the others continue.
func (s Service) ConvertLegacyJobs(ctx context.Context) error {
	logger.Infof("Converting legacy jobs to the split format...")
	jobs, err := s.db.ListLegacyCatalogJobs()
	if err != nil {
		return fmt.Errorf("streams conversion: failed to list legacy jobs: %s", err)
	}

	var supported []*models.Job
	for _, job := range jobs {
		if utils.SupportsStreamsV2(job.Source.Version) {
			supported = append(supported, job)
		}
	}
	err = utils.ForEachConcurrently(ctx, supported, constants.StreamsMigrationConcurrency, func(ctx context.Context, _ int, job *models.Job) error {
		if err := s.convertAndStoreLegacyCatalog(ctx, job); err != nil {
			logger.Warnf("streams conversion: job_id[%d] stays legacy: %s", job.ID, err)
			telemetry.TrackStreamsMigrationFailure(ctx, telemetry.MigrationTriggerStartup, job.Source, 1, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("streams conversion: context cancelled: %s", err)
	}

	logger.Infof("supported legacy jobs converted to the split format")
	return nil
}

// convertAndStoreLegacyCatalog converts one job and stores its split catalog. A job edited or deleted
// meanwhile is skipped, not failed.
func (s Service) convertAndStoreLegacyCatalog(ctx context.Context, job *models.Job) error {
	catalog, err := s.convertLegacyCatalog(ctx, job, job.Source)
	if err != nil {
		return fmt.Errorf("failed to convert streams: %s", err)
	}
	updated, err := s.db.SetStreamsV2Catalog(catalog)
	if err != nil {
		return fmt.Errorf("failed to store the split catalog: %s", err)
	}
	if !updated {
		logger.Infof("streams conversion: job_id[%d] changed during conversion, skipping", job.ID)
		return nil
	}
	logger.Infof("Job %d converted to the split format successfully", job.ID)
	return nil
}

// migrateForSourceUpgrade converts the legacy jobs with source's version,
// constants.StreamsMigrationConcurrency at a time, and returns their split catalogs.
func (s Service) migrateForSourceUpgrade(ctx context.Context, source *models.Source, jobs []*models.Job) ([]database.ConvertedCatalog, error) {
	var legacyJobs []*models.Job
	for _, job := range jobs {
		if !job.StreamsCatalog().IsSplit() {
			legacyJobs = append(legacyJobs, job)
		}
	}
	catalogs := make([]database.ConvertedCatalog, len(legacyJobs))
	err := utils.ForEachConcurrently(ctx, legacyJobs, constants.StreamsMigrationConcurrency, func(ctx context.Context, i int, job *models.Job) error {
		catalog, err := s.convertLegacyCatalog(ctx, job, source)
		if err != nil {
			return fmt.Errorf("job_id[%d] could not be migrated to the split streams format: %s", job.ID, err)
		}
		catalogs[i] = catalog
		return nil
	})
	if err != nil {
		telemetry.TrackStreamsMigrationFailure(ctx, telemetry.MigrationTriggerSourceUpgrade, source, len(legacyJobs), err)
		return nil, fmt.Errorf("source version upgrade aborted: %s", err)
	}
	return catalogs, nil
}

// convertLegacyCatalog converts job's legacy streams_config to the split catalog with source's driver.
func (s Service) convertLegacyCatalog(ctx context.Context, job *models.Job, source *models.Source) (database.ConvertedCatalog, error) {
	legacy := utils.StringValue(job.StreamsConfig)
	catalog, err := s.temporal.ConvertStreams(ctx, job.ID, source.Type, source.Version, legacy)
	if err != nil {
		return database.ConvertedCatalog{}, err
	}
	return database.ConvertedCatalog{JobID: job.ID, From: legacy, Available: catalog.Available, Selected: catalog.Selected}, nil
}
