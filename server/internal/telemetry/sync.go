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
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/internal/telemetry/utils"
)

// jobDetails contains common job information used in telemetry
type jobDetails struct {
	JobName         string
	CreatedAt       time.Time
	CreatedBy       string
	SourceType      string
	SourceName      string
	DestinationType string
	DestinationName string
}

// getJobDetails fetches common job information from the database
func getJobDetails(jobID int) (*jobDetails, error) {
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %v", err)
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	details := &jobDetails{
		JobName:   job.Name,
		CreatedAt: job.CreatedAt,
	}

	if job.CreatedBy != nil {
		userORM := database.NewUserORM()
		fullUser, err := userORM.GetByID(job.CreatedBy.ID)
		if err != nil {
			logs.Error("Failed to get user details for telemetry: %v", err)
			return nil, err
		}
		details.CreatedBy = fullUser.Username
	}

	if job.SourceID != nil {
		details.SourceType = job.SourceID.Type
		details.SourceName = job.SourceID.Name
	}

	if job.DestID != nil {
		details.DestinationType = job.DestID.DestType
		details.DestinationName = job.DestID.Name
	}

	return details, nil
}

// TrackSyncStart tracks when a sync job starts with relevant properties
func TrackSyncStart(ctx context.Context, jobID int, workflowID string) error {
	properties := map[string]interface{}{
		"job_id":      jobID,
		"workflow_id": workflowID,
		"started_at":  time.Now().UTC().Format(time.RFC3339),
	}

	// Get job details from database
	details, err := getJobDetails(jobID)

	if err != nil {
		logs.Error("Failed to get job details for telemetry: %v", err)
		return err
	}
	properties["job_name"] = details.JobName
	properties["created_at"] = details.CreatedAt.Format(time.RFC3339)
	properties["created_by"] = details.CreatedBy
	properties["source_type"] = details.SourceType
	properties["source_name"] = details.SourceName
	properties["destination_type"] = details.DestinationType
	properties["destination_name"] = details.DestinationName

	if err := TrackEvent(ctx, utils.EventSyncStarted, properties); err != nil {
		logs.Error("Failed to track sync start event: %v", err)
		return err
	}

	return nil
}

// TrackSyncFailed tracks when a sync job fails with relevant properties
func TrackSyncFailed(ctx context.Context, jobID int, workflowID string) error {
	properties := map[string]interface{}{
		"job_id":      jobID,
		"workflow_id": workflowID,
		"ended_at":    time.Now().UTC().Format(time.RFC3339),
	}

	// Get job details from database
	details, err := getJobDetails(jobID)
	if err != nil {
		logs.Error("Failed to get job details for telemetry: %v", err)
		return err
	}
	properties["job_name"] = details.JobName
	properties["created_at"] = details.CreatedAt.Format(time.RFC3339)
	properties["created_by"] = details.CreatedBy
	properties["source_type"] = details.SourceType
	properties["source_name"] = details.SourceName
	properties["destination_type"] = details.DestinationType
	properties["destination_name"] = details.DestinationName

	if err := TrackEvent(ctx, utils.EventSyncFailed, properties); err != nil {
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
	details, err := getJobDetails(jobID)
	if err != nil {
		logs.Error("Failed to get job details for telemetry: %v", err)
		return err
	}
	properties["job_name"] = details.JobName
	properties["created_at"] = details.CreatedAt.Format(time.RFC3339)
	properties["created_by"] = details.CreatedBy
	properties["source_type"] = details.SourceType
	properties["source_name"] = details.SourceName
	properties["destination_type"] = details.DestinationType
	properties["destination_name"] = details.DestinationName
	// Read stats.json file
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	homeDir := docker.GetDefaultConfigDir()
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	statsPath := filepath.Join(mainSyncDir, "stats.json")

	statsData, err := os.ReadFile(statsPath)
	if err != nil {
		logs.Error("Failed to read stats.json: %v", err)
		return err
	}
	var stats map[string]interface{}
	if err := json.Unmarshal(statsData, &stats); err != nil {
		logs.Error("Failed to unmarshal stats.json: %v", err)
		return err
	}
	// Add stats properties to the event
	if recordsSynced, ok := stats["Synced Records"]; ok {
		properties["records_synced"] = recordsSynced
	}
	if memory, ok := stats["Memory"]; ok {
		properties["memory_used"] = memory
	}

	// Read streams.json if exists
	streamsPath := filepath.Join(mainSyncDir, "streams.json")
	streamsData, err := os.ReadFile(streamsPath)
	if err != nil {
		logs.Error("Failed to read streams.json: %v", err)
		return err
	}
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

	if err := json.Unmarshal(streamsData, &streamsConfig); err != nil {
		logs.Error("Error unmarshalling streams.json: %v", err)
	}
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

	if err := TrackEvent(ctx, utils.EventSyncCompleted, properties); err != nil {
		logs.Error("Failed to track sync completion event: %v", err)
		return err
	}

	return nil
}
