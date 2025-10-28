package telemetry

import (
	"context"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// TrackJobCreation tracks the creation of a new job with relevant properties
func TrackJobCreation(ctx context.Context, job *models.Job) {
	go func() {
		if instance == nil || job == nil {
			return
		}

		properties := map[string]interface{}{
			"job_id":           job.ID,
			"job_name":         job.Name,
			"project_id":       job.ProjectID,
			"source_type":      job.SourceID.Type,
			"source_name":      job.SourceID.Name,
			"destination_type": job.DestID.DestType,
			"destination_name": job.DestID.Name,
			"frequency":        job.Frequency,
			"active":           job.Active,
		}

		if !job.CreatedAt.IsZero() {
			properties["created_at"] = job.CreatedAt.Format(time.RFC3339)
		}

		if err := TrackEvent(ctx, EventJobCreated, properties); err != nil {
			logs.Debug("Failed to track job creation event: %s", err)
			return
		}
		TrackJobEntity(ctx)
	}()
}

func TrackJobEntity(ctx context.Context) {
	TrackSourcesStatus(ctx)
	TrackDestinationsStatus(ctx)
}
