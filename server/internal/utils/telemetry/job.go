package telemetry

import (
	"context"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// TrackJobCreation tracks the creation of a new job with relevant properties
func TrackJobCreation(ctx context.Context, job *models.Job) {
	go func() {
		if instance == nil || job == nil {
			return
		}

		properties := map[string]interface{}{
			"job_id":     job.ID,
			"job_name":   job.Name,
			"project_id": job.ProjectID,
			"frequency":  job.Frequency,
			"active":     job.Active,
		}

		// Safely add source properties
		if job.Source != nil {
			properties["source_type"] = job.Source.Type
			properties["source_name"] = job.Source.Name
			if job.Source.Version != "" {
				properties["source_olake_version"] = job.Source.Version
			}
		}

		// Safely add destination properties
		if job.Destination != nil {
			properties["destination_type"] = job.Destination.DestType
			properties["destination_name"] = job.Destination.Name
			if job.Destination.Version != "" {
				properties["destination_olake_version"] = job.Destination.Version
			}
		}

		if !job.CreatedAt.IsZero() {
			properties["created_at"] = job.CreatedAt.Format(time.RFC3339)
		}

		if err := TrackEvent(ctx, EventJobCreated, properties); err != nil {
			logger.Debug("Failed to track job creation event: %s", err)
			return
		}
	}()
}
