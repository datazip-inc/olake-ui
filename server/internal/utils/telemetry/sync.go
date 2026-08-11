package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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
	Frequency          string
	Catalog            string
	StreamsConfig      string
}

func getJobDetails(jobID int) (*jobDetails, error) {
	job, err := instance.db.GetJobByID(jobID, true)
	if err != nil {
		// A decryption failure must not cost us the whole event - retry
		// without decrypt; catalog just falls back to "none" below.
		job, err = instance.db.GetJobByID(jobID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get job details: %s", err)
		}
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	details := &jobDetails{
		JobName:       job.Name,
		CreatedAt:     job.CreatedAt,
		Frequency:     job.Frequency,
		StreamsConfig: job.StreamsConfig,
		Catalog:       "unknown",
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
		"frequency":           details.Frequency,
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

		details, err := getJobDetails(info.JobID)
		if err != nil {
			logger.Debugf("failed to track %s event: %s", eventType, err)
			return
		}

		properties := prepareCommonProperties(info, eventType, details)
		addStreamsProperties(properties, details.StreamsConfig)

		// Best-effort: a missing or unparsable stats.json must not drop the
		// event. There's nothing to enrich yet at "started".
		if eventType != EventSyncStarted {
			enrichWithSyncStats(properties, info.WorkflowID)
		}

		if err := TrackEvent(context.Background(), eventType, properties); err != nil {
			logger.Debugf("failed to track %s event: %s", eventType, err)
		}
	}()
}

// enrichWithSyncStats attaches records_synced/memory_used from stats.json when
// available. Silently no-ops if the file is missing or unparsable - callers
// must not treat that as a reason to drop the event.
func enrichWithSyncStats(properties map[string]interface{}, workflowID string) {
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	statsPath := filepath.Join(constants.DefaultConfigDir, syncFolderName, "stats.json")

	statsData, err := os.ReadFile(statsPath)
	if err != nil {
		return
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(statsData, &stats); err != nil {
		return
	}

	if recordsSynced, ok := stats["Synced Records"]; ok {
		properties["records_synced"] = recordsSynced
	}
	if memory, ok := stats["Memory"]; ok {
		properties["memory_used"] = memory
	}
}

// addStreamsProperties computes sync-mode and stream-selection properties
// from the job's DB-backed streams_config, mirroring olake core's
// types.Catalog shape: {selected_streams: {ns: [{stream_name,...}]},
// streams: [{stream: {name, namespace, sync_mode}}]}. Sourcing this from the
// DB rather than the on-disk streams.json means it's available on every sync
// event, not just sync_completed. Emits counts only - never stream names,
// which are customer data.
func addStreamsProperties(properties map[string]interface{}, streamsConfig string) {
	if streamsConfig == "" {
		return
	}

	var cfg struct {
		SelectedStreams map[string][]struct {
			StreamName     string `json:"stream_name"`
			Normalization  bool   `json:"normalization"`
			PartitionRegex string `json:"partition_regex"`
		} `json:"selected_streams"`
		Streams []struct {
			Stream struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				SyncMode  string `json:"sync_mode"`
			} `json:"stream"`
		} `json:"streams"`
	}
	if err := json.Unmarshal([]byte(streamsConfig), &cfg); err != nil {
		return
	}

	selected := make(map[string]bool)
	selectedCount, normalizedCount, partitionedCount := 0, 0, 0
	for ns, streams := range cfg.SelectedStreams {
		for _, s := range streams {
			selected[ns+"."+s.StreamName] = true
			selectedCount++
			if s.Normalization {
				normalizedCount++
			}
			if s.PartitionRegex != "" {
				partitionedCount++
			}
		}
	}

	var fullRefreshCount, incrementalCount, cdcCount, strictCDCCount int
	for _, cs := range cfg.Streams {
		key := cs.Stream.Namespace + "." + cs.Stream.Name
		if !selected[key] {
			continue
		}
		switch cs.Stream.SyncMode {
		case "full_refresh":
			fullRefreshCount++
		case "incremental":
			incrementalCount++
		case "cdc":
			cdcCount++
		case "strict_cdc":
			strictCDCCount++
		}
	}

	properties["full_refresh_streams_count"] = fullRefreshCount
	properties["incremental_streams_count"] = incrementalCount
	properties["cdc_streams_count"] = cdcCount
	properties["strict_cdc_streams_count"] = strictCDCCount
	properties["selected_streams_count"] = selectedCount
	properties["normalized_streams_count"] = normalizedCount
	properties["partitioned_streams_count"] = partitionedCount
}
