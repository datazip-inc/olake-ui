package k8s

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	appConfig "olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"

	"k8s.io/apimachinery/pkg/util/validation"
)

// JobMappingStats contains statistics about job mapping loading
type JobMappingStats struct {
	TotalEntries    int
	ValidEntries    int
	InvalidJobIDs   []string
	InvalidMappings []string
}

// validateJobMapping validates a single job mapping entry
func validateJobMapping(jobIDStr string, nodeLabels map[string]string, stats *JobMappingStats) (int, map[string]string, bool) {
	// Parse and validate JobID
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil || jobID <= 0 {
		stats.InvalidJobIDs = append(stats.InvalidJobIDs, jobIDStr)
		if err != nil {
			logger.Errorf("Invalid JobID format (must be integer): '%s'", jobIDStr)
		} else {
			logger.Errorf("JobID must be positive integer, got: %d", jobID)
		}
		return 0, nil, false
	}

	// Handle null/empty mappings
	if nodeLabels == nil {
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: null mapping", jobID))
		logger.Errorf("JobID %d has null node mapping", jobID)
		return 0, nil, false
	}

	if len(nodeLabels) == 0 {
		logger.Warnf("JobID %d has empty node mapping (will use default scheduling)", jobID)
		return jobID, make(map[string]string), true
	}

	// Validate all labels
	validMapping := make(map[string]string)
	for key, value := range nodeLabels {
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)

		if err := validateLabelPair(jobID, key, value, stats); err != nil {
			return 0, nil, false
		}

		validMapping[key] = value
	}

	return jobID, validMapping, true
}

// validateLabelPair validates a single key-value label pair
func validateLabelPair(jobID int, key, value string, stats *JobMappingStats) error {
	if key == "" {
		err := fmt.Errorf("empty label key")
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	if value == "" {
		err := fmt.Errorf("empty label value for key '%s'", key)
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	if errs := validation.IsQualifiedName(key); len(errs) > 0 {
		err := fmt.Errorf("invalid label key '%s': %v", key, errs)
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
		err := fmt.Errorf("invalid label value '%s' for key '%s': %v", value, key, errs)
		stats.InvalidMappings = append(stats.InvalidMappings, fmt.Sprintf("JobID %d: %s", jobID, err))
		logger.Errorf("JobID %d has %s", jobID, err)
		return err
	}

	return nil
}

// LoadJobMappingFromEnv loads the JobID to node mapping configuration from environment variables
// with enhanced error handling and detailed validation
func LoadJobMappingFromEnv() map[int]map[string]string {
	jobMappingJSON := appConfig.GetEnv("OLAKE_JOB_MAPPING", "")
	if jobMappingJSON == "" {
		logger.Info("No JobID to Node mapping found, using empty mapping")
		return make(map[int]map[string]string)
	}

	var jobMapping map[string]map[string]string
	if err := json.Unmarshal([]byte(jobMappingJSON), &jobMapping); err != nil {
		logger.Errorf("Failed to parse JobID to Node mapping as JSON: %v", err)
		logger.Errorf("Raw configuration: %s", jobMappingJSON)
		return make(map[int]map[string]string)
	}

	// Enhanced validation and conversion with detailed error tracking
	stats := JobMappingStats{
		TotalEntries:    len(jobMapping),
		InvalidJobIDs:   make([]string, 0),
		InvalidMappings: make([]string, 0),
	}

	result := make(map[int]map[string]string)

	for jobIDStr, nodeLabels := range jobMapping {
		if jobID, validMapping, isValid := validateJobMapping(jobIDStr, nodeLabels, &stats); isValid {
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

	if len(stats.InvalidJobIDs) > 0 {
		logger.Errorf("Found %d invalid JobIDs: %v",
			len(stats.InvalidJobIDs), stats.InvalidJobIDs)
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

	return result
}
