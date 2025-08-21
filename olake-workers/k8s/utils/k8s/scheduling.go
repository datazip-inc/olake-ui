package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	appConfig "olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"

	"k8s.io/apimachinery/pkg/util/validation"
)

// Package-level variable to store last known good mapping for fallback
var lastValidMapping map[int]map[string]string

// JobMappingStats contains statistics about job mapping loading
type JobMappingStats struct {
	TotalEntries    int
	ValidEntries    int
	InvalidMappings []string
}

// validateJobMapping validates a single job mapping entry
func validateJobMapping(jobID int, nodeLabels map[string]string, stats *JobMappingStats) (map[string]string, bool) {
	// Validate JobID
	if jobID <= 0 {
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("Invalid JobID: %d", jobID))
		return nil, false
	}

	// Handle null/empty mappings
	if nodeLabels == nil {
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: null mapping", jobID))
		// logger.Errorf("JobID %d has null node mapping", jobID)
		return nil, false
	}

	if len(nodeLabels) == 0 {
		// logger.Warnf("JobID %d has empty node mapping (will use default scheduling)", jobID)
		return make(map[string]string), true
	}

	// Validate all labels
	validMapping := make(map[string]string)
	for key, value := range nodeLabels {
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)

		if err := validateLabelPair(jobID, key, value, stats); err != nil {
			return nil, false
		}

		validMapping[key] = value
	}

	return validMapping, true
}

// validateLabelPair validates a single key-value label pair
func validateLabelPair(jobID int, key, value string, stats *JobMappingStats) error {
	if key == "" {
		err := fmt.Errorf("empty label key")
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		// logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	if value == "" {
		err := fmt.Errorf("empty label value for key '%s'", key)
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		// logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	if errs := validation.IsQualifiedName(key); len(errs) > 0 {
		err := fmt.Errorf("invalid label key '%s': %v", key, errs)
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		// logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
		err := fmt.Errorf("invalid label value '%s' for key '%s': %v", value, key, errs)
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		// logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	return nil
}

// LoadJobMapping loads the JobID to node mapping configuration from config
// with enhanced error handling and detailed validation
func LoadJobMapping(cfg *appConfig.Config) map[int]map[string]string {
	// Use only config-provided mapping; no direct env reads
	if cfg.Kubernetes.JobMapping == nil {
		logger.Info("No JobID to Node mapping found in config, using empty mapping")
		return make(map[int]map[string]string)
	}

	// Enhanced validation and conversion with detailed error tracking
	stats := JobMappingStats{
		TotalEntries:    len(cfg.Kubernetes.JobMapping),
		InvalidMappings: make([]string, 0),
	}

	result := make(map[int]map[string]string)

	for jobID, nodeLabels := range cfg.Kubernetes.JobMapping {
		if validMapping, ok := validateJobMapping(jobID, nodeLabels, &stats); ok {
			result[jobID] = validMapping
			stats.ValidEntries++
		}
	}

	// Log comprehensive statistics
	logger.Infof("Job mapping loaded: %d valid entries out of %d total",
		stats.ValidEntries, stats.TotalEntries)

	// Print the valid job mapping configuration as JSON
	if len(result) > 0 {
		if jsonBytes, err := json.Marshal(result); err == nil {
			logger.Infof("Job mapping configuration: %s", string(jsonBytes))
		}
	}

	if len(stats.InvalidMappings) > 0 {
		logger.Errorf("Found %d invalid mappings: %v",
			len(stats.InvalidMappings), stats.InvalidMappings)
	}

	// Warn if no valid mappings were loaded
	if stats.ValidEntries == 0 && stats.TotalEntries > 0 {
		logger.Errorf("No valid job mappings loaded despite %d entries in configuration",
			stats.TotalEntries)
	}

	// Fallback to last valid mapping if available
	if stats.ValidEntries == 0 && lastValidMapping != nil {
		logger.Debugf("Falling back to previous valid mapping with %d entries", len(lastValidMapping))
		return lastValidMapping
	}

	// Store successful result as fallback for future failures
	if len(result) > 0 || stats.ValidEntries > 0 {
		lastValidMapping = result
		logger.Debugf("Cached valid mapping with %d entries for future fallback", len(result))
	}

	return result
}
