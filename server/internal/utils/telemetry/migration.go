package telemetry

import (
	"context"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// Streams V1 -> V2 migration triggers
const (
	MigrationTriggerStartup       = "startup"
	MigrationTriggerSourceUpgrade = "source_upgrade"
)

// TrackStreamsMigrationFailure tracks a Streams V1 -> V2 migration that left jobCount jobs of source on V1
func TrackStreamsMigrationFailure(ctx context.Context, trigger string, source *models.Source, jobCount int, migrationErr error) {
	go func() {
		if instance == nil || source == nil || migrationErr == nil {
			return
		}

		properties := map[string]interface{}{
			"trigger":        trigger,
			"source_id":      source.ID,
			"source_type":    source.Type,
			"source_version": source.Version,
			"job_count":      jobCount,
			"error":          migrationErr.Error(),
		}

		if err := TrackEvent(ctx, EventStreamsMigrationFailed, properties); err != nil {
			logger.Debugf("failed to track streams migration failure event: %s", err)
		}
	}()
}
