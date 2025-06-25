package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
)

// TrackSyncStart tracks when a sync job starts with relevant properties
func TrackSyncStart(ctx context.Context, jobID int, workflowID string) error {
	properties := map[string]interface{}{
		"job_id":      jobID,
		"workflow_id": workflowID,
		"started_at":  time.Now().UTC().Format(time.RFC3339),
	}

	// Get job details from database
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID)
	if err != nil {
		logs.Error("Failed to get job details for telemetry: %v", err)
	} else if job != nil {
		properties["job_name"] = job.Name
		properties["created_at"] = job.CreatedAt.Format(time.RFC3339)

		if job.SourceID != nil {
			properties["source_type"] = job.SourceID.Type
			properties["source_name"] = job.SourceID.Name
		}

		if job.DestID != nil {
			properties["destination_type"] = job.DestID.DestType
			properties["destination_name"] = job.DestID.Name
		}
	}

	if err := TrackEvent(ctx, constants.EventSyncStarted, properties); err != nil {
		logs.Error("Failed to track sync start event: %v", err)
		return err
	}

	return nil
}

// TrackSyncFailed tracks when a sync job fails with relevant properties
func TrackSyncFailed(ctx context.Context, jobID int, workflowID, jobName string, createdAt time.Time, sourceType, sourceName, destinationType, destinationName string) error {
	properties := map[string]interface{}{
		"job_id":           jobID,
		"workflow_id":      workflowID,
		"ended_at":         time.Now().UTC().Format(time.RFC3339),
		"job_name":         jobName,
		"created_at":       createdAt,
		"source_type":      sourceType,
		"source_name":      sourceName,
		"destination_type": destinationType,
		"destination_name": destinationName,
	}

	if err := TrackEvent(ctx, constants.EventSyncFailed, properties); err != nil {
		logs.Error("Failed to track sync failure event: %v", err)
		return err
	}

	return nil
}

// TrackSyncCompleted tracks when a sync job completes successfully with relevant properties
func TrackSyncCompleted(ctx context.Context, jobID int, workflowID string) error {
	properties := map[string]interface{}{
		"job_id":      jobID,
		"workflow_id": workflowID,
		"ended_at":    time.Now().UTC().Format(time.RFC3339),
	}

	// Get job details from database
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID)
	if err != nil {
		logs.Error("Failed to get job details for telemetry: %v", err)
	} else if job != nil {
		properties["job_name"] = job.Name
		properties["created_at"] = job.CreatedAt.Format(time.RFC3339)

		if job.CreatedBy != nil {
			userORM := database.NewUserORM()
			if fullUser, err := userORM.GetByID(job.CreatedBy.ID); err == nil {
				properties["created_by"] = fullUser.Username
			}
		}

		if job.SourceID != nil {
			properties["source_type"] = job.SourceID.Type
			properties["source_name"] = job.SourceID.Name
		}

		if job.DestID != nil {
			properties["destination_type"] = job.DestID.DestType
			properties["destination_name"] = job.DestID.Name
		}
	}

	// Read stats.json file
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	homeDir := docker.GetDefaultConfigDir()
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	statsPath := filepath.Join(mainSyncDir, "stats.json")

	if statsData, err := os.ReadFile(statsPath); err == nil {
		var stats map[string]interface{}
		if err := json.Unmarshal(statsData, &stats); err == nil {
			// Add stats properties to the event
			if recordsSynced, ok := stats["Synced Records"]; ok {
				properties["records_synced"] = recordsSynced
			}
			if memory, ok := stats["Memory"]; ok {
				properties["memory_used"] = memory
			}
		}
		if err != nil {
			logs.Error("Error unmarshalling stats.json: %v", err)
		}
	}

	// Read streams.json if exists
	streamsPath := filepath.Join(mainSyncDir, "streams.json")
	if streamsData, err := os.ReadFile(streamsPath); err == nil {
		var streamsConfig struct {
			Streams []struct {
				Stream struct {
					Name               string   `json:"name"`
					Namespace          string   `json:"namespace"`
					SyncMode           string   `json:"sync_mode"`
					SupportedSyncModes []string `json:"supported_sync_modes"`
				} `json:"stream"`
			} `json:"streams"`
			SelectedStreams map[string][]struct {
				StreamName     string `json:"stream_name"`
				Normalization  bool   `json:"normalization"`
				PartitionRegex string `json:"partition_regex"`
			} `json:"selected_streams"`
		}

		if err := json.Unmarshal(streamsData, &streamsConfig); err == nil {
			// Count normalized streams
			normalizedCount := 0
			partitionedCount := 0

			// Count normalized and partitioned streams from selected_streams
			for _, streams := range streamsConfig.SelectedStreams {
				for _, stream := range streams {
					if stream.Normalization {
						normalizedCount++
					}
					if stream.PartitionRegex != "" {
						partitionedCount++
					}
				}
			}

			properties["normalized_streams_count"] = normalizedCount
			properties["partitioned_streams_count"] = partitionedCount
		} else {
			logs.Error("Error unmarshalling streams.json: %v", err)
		}
	}

	if err := TrackEvent(ctx, constants.EventSyncCompleted, properties); err != nil {
		logs.Error("Failed to track sync completion event: %v", err)
		return err
	}

	return nil
}
