package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

type jobDetails struct {
	JobName         string
	CreatedAt       time.Time
	CreatedBy       string
	SourceType      string
	SourceName      string
	DestinationType string
	DestinationName string
}

func getJobDetails(jobID int) (*jobDetails, error) {
	job, err := instance.db.GetJobByID(jobID, false)
	if err != nil || job == nil {
		if job == nil {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job details: %s", err)
	}

	details := &jobDetails{
		JobName:   job.Name,
		CreatedAt: job.CreatedAt,
	}

	if job.CreatedBy != nil {
		if user, err := instance.db.GetUserByID(job.CreatedBy.ID); err == nil {
			details.CreatedBy = user.Username
		}
	}

	if job.Source != nil {
		details.SourceType = job.Source.Type
		details.SourceName = job.Source.Name
	}

	if job.Destination != nil {
		details.DestinationType = job.Destination.DestType
		details.DestinationName = job.Destination.Name
	}

	return details, nil
}

func prepareCommonProperties(jobID int, workflowID, executionEnvironment string, details *jobDetails, eventTime time.Time) map[string]interface{} {
	props := map[string]interface{}{
		"job_id":           jobID,
		"workflow_id":      workflowID,
		"job_name":         details.JobName,
		"created_at":       details.CreatedAt.Format(time.RFC3339),
		"created_by":       details.CreatedBy,
		"source_type":      details.SourceType,
		"source_name":      details.SourceName,
		"destination_type": details.DestinationType,
		"destination_name": details.DestinationName,
		"environment":      executionEnvironment,
	}

	if eventTime.IsZero() {
		eventTime = time.Now().UTC()
	}
	timeKey := "started_at"
	if eventTime != props["created_at"] {
		timeKey = "ended_at"
	}
	props[timeKey] = eventTime.Format(time.RFC3339)

	return props
}

func trackSyncEvent(ctx context.Context, jobID int, workflowID, executionEnvironment, eventType string) error {
	details, err := getJobDetails(jobID)
	if err != nil {
		return err
	}

	properties := prepareCommonProperties(jobID, workflowID, executionEnvironment, details, time.Time{})
	if eventType == EventSyncCompleted {
		if err := enrichWithSyncStats(ctx, properties, workflowID); err != nil {
			return err
		}
	}

	if err := TrackEvent(ctx, eventType, properties); err != nil {
		return err
	}
	return nil
}

func enrichWithSyncStats(ctx context.Context, properties map[string]interface{}, workflowID string) error {
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	statsData, err := readSyncJobFile(ctx, syncFolderName, "stats.json")
	if err != nil {
		return err
	}
	if err := addStatsProperties(properties, statsData); err != nil {
		return err
	}

	streamsData, err := readSyncJobFile(ctx, syncFolderName, "streams.json")
	if err != nil {
		return err
	}

	return addStreamsProperties(properties, streamsData)
}

func readSyncJobFile(ctx context.Context, syncFolderName, filename string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode)) {
	case constants.StorageModeS3:
		relPath := path.Join(syncFolderName, filename)
		return storage.ReadFileFromS3AtRelPath(ctx, relPath)
	case constants.StorageModeNFS:
		filePath := filepath.Join(constants.DefaultConfigDir, syncFolderName, filename)
		return os.ReadFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported OLAKE_STORAGE_MODE: %q", appconfig.Load().OlakeStorageMode)
	}
}

func addStatsProperties(properties map[string]interface{}, statsData []byte) error {
	var stats map[string]interface{}
	if err := json.Unmarshal(statsData, &stats); err != nil {
		return err
	}

	if recordsSynced, ok := stats["Synced Records"]; ok {
		properties["records_synced"] = recordsSynced
	}
	if memory, ok := stats["Memory"]; ok {
		properties["memory_used"] = memory
	} else if memory, ok := stats["Memory Usage Bytes"]; ok {
		properties["memory_used"] = memory
	}
	return nil
}

func addStreamsProperties(properties map[string]interface{}, streamsData []byte) error {

	var streamsConfig struct {
		SelectedStreams map[string][]struct {
			Normalization  bool   `json:"normalization"`
			PartitionRegex string `json:"partition_regex"`
		} `json:"selected_streams"`
	}

	if err := json.Unmarshal(streamsData, &streamsConfig); err != nil {
		return fmt.Errorf("error unmarshalling streams.json: %s", err)
	}

	normalizedCount, partitionedCount := 0, 0
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
	return nil
}

func TrackSyncStart(ctx context.Context, jobID int, workflowID, executionEnvironment string) {
	go func() {
		if instance == nil {
			return
		}

		err := trackSyncEvent(ctx, jobID, workflowID, executionEnvironment, EventSyncStarted)
		if err != nil {
			logger.Debug("failed to track sync start event: %s", err)
		}
	}()
}

func TrackSyncFailed(jobID int, workflowID, executionEnvironment string) {
	go func() {
		if instance == nil {
			return
		}

		err := trackSyncEvent(context.Background(), jobID, workflowID, executionEnvironment, EventSyncFailed)
		if err != nil {
			logger.Debug("failed to track sync failed event: %s", err)
		}
	}()
}

func TrackSyncCompleted(jobID int, workflowID, executionEnvironment string) {
	go func() {
		if instance == nil {
			return
		}

		err := trackSyncEvent(context.Background(), jobID, workflowID, executionEnvironment, EventSyncCompleted)
		if err != nil {
			logger.Debug("failed to track sync completed event: %s", err)
		}
	}()
}
