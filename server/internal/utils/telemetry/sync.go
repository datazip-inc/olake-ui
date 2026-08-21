package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// SyncEventInfo carries everything needed to track a sync event, forwarded
// by the worker's sync-telemetry callback.
type SyncEventInfo struct {
	JobID                int
	WorkflowID           string
	ExecutionEnvironment string
	SyncRunCount         int
	Properties           map[string]any
}

type jobDetails struct {
	JobName            string
	CreatedAt          time.Time
	CreatedBy          string
	SourceType         string
	SourceName         string
	SourceVersion      string
	DestinationType    string
	DestinationName    string
	DestinationVersion string
	Catalog            string
}

func getJobDetails(jobID int) (*jobDetails, error) {
	job, err := instance.db.GetJobByID(jobID, true)
	if errors.Is(err, constants.ErrConfigDecrypt) {
		// Decryption failure must not cost us the whole event - retry
		// without decrypt; catalog falls back to "unknown" below.
		job, err = instance.db.GetJobByID(jobID, false)
	}
	if err != nil || job == nil {
		if job == nil {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job details: %s", err)
	}

	details := &jobDetails{
		JobName:   job.Name,
		CreatedAt: job.CreatedAt,
		Catalog:   "unknown",
	}

	if job.CreatedBy != nil {
		if user, err := instance.db.GetUserByID(job.CreatedBy.ID); err == nil {
			details.CreatedBy = user.Username
		}
	}

	if job.Source != nil {
		details.SourceType = job.Source.Type
		details.SourceName = job.Source.Name
		details.SourceVersion = job.Source.Version
	}

	if job.Destination != nil {
		details.DestinationType = job.Destination.DestType
		details.DestinationName = job.Destination.Name
		details.DestinationVersion = job.Destination.Version
		details.Catalog = parseCatalogType(job.Destination.Config)
	}

	return details, nil
}

// parseCatalogType extracts config.writer.catalog_type from a destination
// config JSON string. "none" means the config has no catalog_type (e.g.
// Parquet); "unknown" means it couldn't be determined (unparsable config).
func parseCatalogType(config string) string {
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return "unknown"
	}
	if writer, ok := configMap["writer"].(map[string]interface{}); ok {
		if catalogType, ok := writer["catalog_type"].(string); ok && catalogType != "" {
			return catalogType
		}
	}
	return "none"
}

func prepareCommonProperties(info SyncEventInfo, eventType string, details *jobDetails) map[string]interface{} {
	props := map[string]interface{}{
		"job_id":              info.JobID,
		"workflow_id":         info.WorkflowID,
		"job_name":            details.JobName,
		"created_at":          details.CreatedAt.Format(time.RFC3339),
		"created_by":          details.CreatedBy,
		"source_type":         details.SourceType,
		"source_name":         details.SourceName,
		"source_version":      details.SourceVersion,
		"destination_type":    details.DestinationType,
		"destination_name":    details.DestinationName,
		"destination_version": details.DestinationVersion,
		"catalog":             details.Catalog,
		"environment":         info.ExecutionEnvironment,
	}

	if info.SyncRunCount > 0 {
		props["sync_run_count"] = info.SyncRunCount
	}
	timeKey := "ended_at"
	if eventType == EventSyncStarted {
		timeKey = "started_at"
	}
	props[timeKey] = time.Now().UTC().Format(time.RFC3339)

	return props
}

// TrackSyncEvent sends a sync event (EventSyncStarted/Completed/Failed/Cancelled)
func TrackSyncEvent(info SyncEventInfo, eventType string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Debugf("recovered panic tracking %s: %v", eventType, r)
			}
		}()
		if instance == nil {
			return
		}

		properties, err := buildProperties(info, eventType)
		if err != nil {
			logger.Debugf("failed to track %s event: %s", eventType, err)
			return
		}

		// Best-effort: a missing or unparsable stats.json/streams.json must
		// not drop the event. There's nothing to enrich yet at "started".
		if eventType != EventSyncStarted {
			if err := enrichWithSyncStats(properties, info.WorkflowID); err != nil {
				logger.Debugf("failed to enrich %s event: %s", eventType, err)
			}
		}

		if err := TrackEvent(context.Background(), eventType, properties); err != nil {
			logger.Debugf("failed to track %s event: %s", eventType, err)
		}
	}()
}

// buildProperties forwards the worker's pre-resolved properties when present.
// Falls back to enriching from the DB for legacy workers that don't send them.
func buildProperties(info SyncEventInfo, eventType string) (map[string]interface{}, error) {
	if len(info.Properties) > 0 {
		props := make(map[string]interface{}, len(info.Properties)+1)
		for k, v := range info.Properties {
			props[k] = v
		}
		props["workflow_id"] = info.WorkflowID

		timeKey := "ended_at"
		if eventType == EventSyncStarted {
			timeKey = "started_at"
		}
		props[timeKey] = time.Now().UTC().Format(time.RFC3339)

		if details, err := getJobDetails(info.JobID); err == nil {
			props["catalog"] = details.Catalog
		}
		return props, nil
	}

	// Old worker - upgraded workers always send Properties, so this path is dead post-upgrade.
	details, err := getJobDetails(info.JobID)
	if err != nil {
		return nil, err
	}
	return prepareCommonProperties(info, eventType, details), nil
}

func enrichWithSyncStats(properties map[string]interface{}, workflowID string) error {
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	mainSyncDir := filepath.Join(constants.DefaultConfigDir, syncFolderName)

	if err := addStatsProperties(properties, mainSyncDir); err != nil {
		return err
	}

	return addStreamsProperties(properties, mainSyncDir)
}

func addStatsProperties(properties map[string]interface{}, mainSyncDir string) error {
	statsPath := filepath.Join(mainSyncDir, "stats.json")
	statsData, err := os.ReadFile(statsPath)
	if err != nil {
		return err
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(statsData, &stats); err != nil {
		return err
	}

	if recordsSynced, ok := stats["Synced Records"]; ok {
		properties["records_synced"] = recordsSynced
	}
	if memory, ok := stats["Memory"]; ok {
		properties["memory_used"] = memory
	}
	return nil
}

func addStreamsProperties(properties map[string]interface{}, mainSyncDir string) error {
	streamsPath := filepath.Join(mainSyncDir, "streams.json")
	streamsData, err := os.ReadFile(streamsPath)
	if err != nil {
		return fmt.Errorf("failed to read streams.json: %s", err)
	}

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
