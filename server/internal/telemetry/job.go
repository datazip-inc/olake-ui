package telemetry

import (
	"context"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/telemetry/utils"
)

// TrackJobCreation tracks the creation of a new job with relevant properties
func TrackJobCreation(ctx context.Context, jobID int, jobName, projectID, sourceType, sourceName, destinationType, destinationName, frequency string, active bool, userID interface{}, createdAt time.Time) error {
	properties := map[string]interface{}{
		"job_id":           jobID,
		"job_name":         jobName,
		"project_id":       projectID,
		"source_type":      sourceType,
		"source_name":      sourceName,
		"destination_type": destinationType,
		"destination_name": destinationName,
		"frequency":        frequency,
		"active":           active,
	}

	if userID != nil {
		properties["user_id"] = userID
	}

	if !createdAt.IsZero() {
		properties["created_at"] = createdAt.Format(time.RFC3339)
	}

	if err := TrackEvent(ctx, utils.EventJobCreated, properties); err != nil {
		logs.Error("Failed to track job creation event: %v", err)
		return err
	}

	return nil
}

// TrackSourcesAndDestinationsStatus logs telemetry about active and inactive sources and destinations
func TrackSourcesAndDestinationsStatus(ctx context.Context, userID interface{}) error {
	sourceORM := database.NewSourceORM()
	destORM := database.NewDestinationORM()
	jobORM := database.NewJobORM()
	userORM := database.NewUserORM()

	// Track sources status
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

	// Track destinations status
	destinations, err := destORM.GetAll()
	if err != nil {
		return err
	}

	activeDestinations := 0
	inactiveDestinations := 0

	for _, dest := range destinations {
		jobs, err := jobORM.GetByDestinationID(dest.ID)
		if err != nil {
			return err
		}
		if len(jobs) > 0 {
			activeDestinations++
		} else {
			inactiveDestinations++
		}
	}

	// Get user properties if available
	var userProps map[string]interface{}
	if userID != nil {
		user, err := userORM.GetByID(userID.(int))
		if err != nil {
			logs.Error("Failed to get user details for telemetry: %v", err)
			return err
		}
		userProps = map[string]interface{}{
			"user_id":    user.ID,
			"user_email": user.Email,
		}
	}

	// Track sources status
	sourceProps := map[string]interface{}{
		"active_sources":   activeSources,
		"inactive_sources": inactiveSources,
		"total_sources":    activeSources + inactiveSources,
	}
	for k, v := range userProps {
		sourceProps[k] = v
	}
	if err := TrackEvent(ctx, utils.EventSourcesUpdated, sourceProps); err != nil {
		return err
	}

	// Track destinations status
	destProps := map[string]interface{}{
		"active_destinations":   activeDestinations,
		"inactive_destinations": inactiveDestinations,
		"total_destinations":    activeDestinations + inactiveDestinations,
	}
	for k, v := range userProps {
		destProps[k] = v
	}
	return TrackEvent(ctx, utils.EventDestinationsUpdated, destProps)
}
