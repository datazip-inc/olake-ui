package telemetry

import (
	"context"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/telemetry/utils"
)

// TrackSourceCreation tracks the creation of a new source with relevant properties
func TrackSourceCreation(ctx context.Context, sourceID int, sourceName, sourceType, version string, createdAt time.Time) error {
	properties := map[string]interface{}{
		"source_id":   sourceID,
		"source_name": sourceName,
		"source_type": sourceType,
		"version":     version,
	}

	if !createdAt.IsZero() {
		properties["created_at"] = createdAt.Format(time.RFC3339)
	}

	if err := TrackEvent(ctx, utils.EventSourceCreated, properties); err != nil {
		logs.Error("Failed to track source creation event: %v", err)
		return err
	}

	return nil
}

// TrackSourcesStatus logs telemetry about active and inactive sources
func TrackSourcesStatus(ctx context.Context, userID interface{}) error {
	sourceORM := database.NewSourceORM()
	jobORM := database.NewJobORM()
	userORM := database.NewUserORM()

	sources, err := sourceORM.GetAll()
	if err != nil {
		return err
	}

	activeSources := 0
	inactiveSources := 0

	for _, source := range sources {
		jobs, err := jobORM.GetBySourceID(source.ID)
		if err != nil {
			return err
		}
		if len(jobs) > 0 {
			activeSources++
		} else {
			inactiveSources++
		}
	}

	// Get user properties if available
	var userProps map[string]interface{}
	if userID != nil {
		if user, err := userORM.GetByID(userID.(int)); err == nil {
			userProps = map[string]interface{}{
				"user_id":    user.ID,
				"user_email": user.Email,
			}
		}
	}

	// Prepare telemetry properties
	props := map[string]interface{}{
		"active_sources":   activeSources,
		"inactive_sources": inactiveSources,
		"total_sources":    activeSources + inactiveSources,
	}

	for k, v := range userProps {
		props[k] = v
	}

	return TrackEvent(ctx, utils.EventSourcesUpdated, props)
}
