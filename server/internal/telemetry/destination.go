package telemetry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
)

// TrackDestinationCreation tracks the creation of a new destination with relevant properties
func TrackDestinationCreation(ctx context.Context, dest models.Destination) {
	go func() {
		properties := map[string]interface{}{
			"destination_id":   dest.ID,
			"destination_name": dest.Name,
			"destination_type": dest.DestType,
			"version":          dest.Version,
			"catalog":          "none",
		}
		var configMap map[string]interface{}
		// parse config to get catalog_type
		if err := json.Unmarshal([]byte(dest.Config), &configMap); err != nil {
			logs.Debug("Failed to unmarshal config: %s", err)
			return
		}

		if writer, exists := configMap["writer"].(map[string]interface{}); exists {
			if catalogType, exists := writer["catalog_type"]; exists {
				properties["catalog"] = catalogType
			}
		}

		if !dest.CreatedAt.IsZero() {
			properties["created_at"] = dest.CreatedAt.Format(time.RFC3339)
		}

		if err := TrackEvent(ctx, EventDestinationCreated, properties); err != nil {
			logs.Debug("Failed to track destination creation event: %s", err)
			return
		}

		// Track destinations status after creation
		TrackDestinationsStatus(ctx)

	}()
}

// TrackDestinationsStatus logs telemetry about active and inactive destinations
func TrackDestinationsStatus(ctx context.Context) {
	go func() {
		// TODO: remove creation of orm from here
		destORM := database.NewDestinationORM()
		jobORM := database.NewJobORM()

		destinations, err := destORM.GetAll()
		if err != nil {
			logs.Debug("Failed to get all destinations: %s", err)
			return
		}

		activeDestinations := 0

		for _, dest := range destinations {
			// TODO: remove db calls loop
			jobs, err := jobORM.GetByDestinationID(dest.ID)
			if err != nil {
				logs.Debug("Failed to get jobs for destination %d: %s", dest.ID, err)
				break
			}
			if len(jobs) > 0 {
				activeDestinations++
			}
		}

		// Prepare telemetry properties
		props := map[string]interface{}{
			"active_destinations":   activeDestinations,
			"inactive_destinations": len(destinations) - activeDestinations,
			"total_destinations":    len(destinations),
		}

		if err := TrackEvent(ctx, EventDestinationsUpdated, props); err != nil {
			logs.Debug("failed to track destination status event: %s", err)
		}
	}()
}
