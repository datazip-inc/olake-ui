package telemetry

import (
	"context"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
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
		if job.SourceID != nil {
			properties["source_type"] = job.SourceID.Type
			properties["source_name"] = job.SourceID.Name
			if job.SourceID.Version != "" {
				properties["source_olake_version"] = job.SourceID.Version
			}
		}

		// Safely add destination properties
		if job.DestID != nil {
			properties["destination_type"] = job.DestID.DestType
			properties["destination_name"] = job.DestID.Name
			if job.DestID.Version != "" {
				properties["destination_olake_version"] = job.DestID.Version
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
