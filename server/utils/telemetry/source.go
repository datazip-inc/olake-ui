package telemetry

import (
	"context"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// TrackSourceCreation tracks the creation of a new source with relevant properties
func TrackSourceCreation(ctx context.Context, source *models.Source) {
	go func() {
		if instance == nil || source == nil {
			return
		}

		properties := map[string]interface{}{
			"source_id":   source.ID,
			"source_name": source.Name,
			"source_type": source.Type,
			"version":     source.Version,
		}

		if !source.CreatedAt.IsZero() {
			properties["created_at"] = source.CreatedAt.Format(time.RFC3339)
		}

		if err := TrackEvent(ctx, EventSourceCreated, properties); err != nil {
			logs.Debug("Failed to track source creation event: %s", err)
			return
		}
		// Track sources status after creation
		TrackSourcesStatus(ctx)
	}()
}

// TrackSourcesStatus logs telemetry about active and inactive sources
func TrackSourcesStatus(ctx context.Context) {
	go func() {
		if instance == nil {
			return
		}

		sources, err := instance.db.ListSources()
		if err != nil {
			logs.Debug("failed to get all sources in track source status: %s", err)
			return
		}

		activeSources := 0
		for _, source := range sources {
			// TODO: remove orm calls from loop
			jobs, err := instance.db.GetJobsBySourceID([]int{source.ID})
			if err != nil {
				logs.Debug("failed to get all jobs for source[%d] in track source status: %s", source.ID, err)
				break
			}
			if len(jobs) > 0 {
				activeSources++
			}
		}

		// Prepare telemetry properties
		props := map[string]interface{}{
			"active_sources":   activeSources,
			"inactive_sources": len(sources) - activeSources,
			"total_sources":    len(sources),
		}

		if err := TrackEvent(ctx, EventSourcesUpdated, props); err != nil {
			logs.Debug("failed to track source status event: %s", err)
		}
	}()
}
