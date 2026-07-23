package etl

import (
	"fmt"
	"path/filepath"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// syncStats holds the pass-through values read from a sync run's stats.json.
// Key contract with the sync CLI: "Synced Records" (number), "Bytes Read"
// (number), "CPU Utilization" (number 0-1), "Memory Usage Bytes" (number).
type syncStats struct {
	Records  float64
	Bytes    float64
	CPURatio float64
	Memory   float64
}

// readSyncStats parses a sync run's stats.json. Missing or unparsable keys yield
// zero values without error (older images ship fewer keys); a missing file or
// truncated JSON (the CLI truncate-rewrites every ~2s) errors so the caller keeps
// the job's previous values.
func readSyncStats(baseDir, workflowID string) (*syncStats, error) {
	statsPath := filepath.Join(baseDir, temporal.GetWorkflowDirectory(temporal.Sync, workflowID), "stats.json")
	raw, err := temporal.ReadJSONFile(statsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", statsPath, err)
	}

	stats := &syncStats{}
	if v, ok := raw["Synced Records"].(float64); ok {
		stats.Records = v
	}
	if v, ok := raw["Bytes Read"].(float64); ok {
		stats.Bytes = v
	}
	if v, ok := raw["CPU Utilization"].(float64); ok {
		stats.CPURatio = v
	}
	if v, ok := raw["Memory Usage Bytes"].(float64); ok {
		stats.Memory = v
	}

	return stats, nil
}
