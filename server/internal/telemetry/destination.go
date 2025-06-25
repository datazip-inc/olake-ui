package telemetry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
)

// TrackDestinationCreation tracks the creation of a new destination with relevant properties
func TrackDestinationCreation(ctx context.Context, destinationID int, destinationName, destinationType, version, config string, createdAt time.Time) error {
	properties := map[string]interface{}{
		"destination_id":   destinationID,
		"destination_name": destinationName,
		"destination_type": destinationType,
		"version":          version,
	}

	properties["catalog"] = "none"
	var configMap map[string]interface{}
	// parse config to get catalog_type
	if err := json.Unmarshal([]byte(config), &configMap); err == nil {
		if writer, exists := configMap["writer"].(map[string]interface{}); exists {
			if catalogType, exists := writer["catalog_type"]; exists {
				properties["catalog"] = catalogType
			}
		}
	}

	if !createdAt.IsZero() {
		properties["created_at"] = createdAt.Format(time.RFC3339)
	}

	if err := TrackEvent(ctx, constants.EventDestinationCreated, properties); err != nil {
		logs.Error("Failed to track destination creation event: %v", err)
		return err
	}

	return nil
}

// TrackDestinationsStatus logs telemetry about active and inactive destinations
func TrackDestinationsStatus(ctx context.Context, userID interface{}) error {
	destORM := database.NewDestinationORM()
	jobORM := database.NewJobORM()
	userORM := database.NewUserORM()

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
		if user, err := userORM.GetByID(userID.(int)); err == nil {
			userProps = map[string]interface{}{
				"user_id":    user.ID,
				"user_email": user.Email,
			}
		}
	}

	// Prepare telemetry properties
	props := map[string]interface{}{
		"active_destinations":   activeDestinations,
		"inactive_destinations": inactiveDestinations,
		"total_destinations":    activeDestinations + inactiveDestinations,
	}

	for k, v := range userProps {
		props[k] = v
	}

	return TrackEvent(ctx, constants.EventDestinationsUpdated, props)
}
