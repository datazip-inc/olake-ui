package etl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// readSyncStats reads and parses stats.json from a sync run's work directory.
// Missing or unparsable keys yield zero values without error (old connector
// images ship fewer keys). A missing file or truncated JSON (the CLI
// truncate-rewrites the file every ~2s) returns an error so the caller can keep
// the previously reported values for that job.
func readSyncStats(dir string) (*syncStats, error) {
	data, err := os.ReadFile(filepath.Join(dir, "stats.json"))
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse stats.json in %s: %s", dir, err)
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
