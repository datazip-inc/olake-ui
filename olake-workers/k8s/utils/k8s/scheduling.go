package k8s

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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

// LoadJobMappingFromEnv loads the JobID to node mapping configuration from environment variables
// with enhanced error handling and detailed validation
func LoadJobMappingFromEnv() map[int]map[string]string {
	jobMappingJSON := os.Getenv("OLAKE_JOB_MAPPING")
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
		// Validate JobID format
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			stats.InvalidJobIDs = append(stats.InvalidJobIDs, jobIDStr)
			logger.Errorf("Invalid JobID format (must be integer): '%s'", jobIDStr)
			continue
		}

		// Validate jobID is positive
		if jobID <= 0 {
			stats.InvalidJobIDs = append(stats.InvalidJobIDs, jobIDStr)
			logger.Errorf("JobID must be positive integer, got: %d", jobID)
			continue
		}

		// Validate node labels structure
		if nodeLabels == nil {
			stats.InvalidMappings = append(stats.InvalidMappings,
				fmt.Sprintf("JobID %d: null mapping", jobID))
			logger.Errorf("JobID %d has null node mapping", jobID)
			continue
		}

		if len(nodeLabels) == 0 {
			logger.Warnf("JobID %d has empty node mapping (will use default scheduling)", jobID)
			result[jobID] = make(map[string]string)
			stats.ValidEntries++
			continue
		}

		// Validate node label keys and values using Kubernetes validation
		// Treat each JobID mapping as atomic: either all labels are valid or reject the entire entry
		validMapping := make(map[string]string)
		isEntryValid := true

		for key, value := range nodeLabels {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)

			// Validate empty key/value
			if key == "" {
				stats.InvalidMappings = append(stats.InvalidMappings,
					fmt.Sprintf("JobID %d: empty label key", jobID))
				logger.Errorf("JobID %d has empty node label key", jobID)
				isEntryValid = false
				break
			}
			if value == "" {
				stats.InvalidMappings = append(stats.InvalidMappings,
					fmt.Sprintf("JobID %d: empty label value for key '%s'", jobID, key))
				logger.Errorf("JobID %d has empty value for node label key '%s'", jobID, key)
				isEntryValid = false
				break
			}

			// Validate Kubernetes label key (qualified name format)
			if errs := validation.IsQualifiedName(key); len(errs) > 0 {
				stats.InvalidMappings = append(stats.InvalidMappings,
					fmt.Sprintf("JobID %d: invalid label key '%s'", jobID, key))
				logger.Errorf("JobID %d has invalid node label key '%s': %v", jobID, key, errs)
				isEntryValid = false
				break
			}

			// Validate Kubernetes label value (RFC 1123 label format)
			if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
				stats.InvalidMappings = append(stats.InvalidMappings,
					fmt.Sprintf("JobID %d: invalid label value '%s' for key '%s'", jobID, value, key))
				logger.Errorf("JobID %d has invalid node label value '%s' for key '%s': %v", jobID, value, key, errs)
				isEntryValid = false
				break
			}

			validMapping[key] = value
		}

		if isEntryValid {
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
