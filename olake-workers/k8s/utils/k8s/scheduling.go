package k8s

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"olake-ui/olake-workers/k8s/logger"
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
		logger.Info("No OLAKE_JOB_MAPPING environment variable found, using empty mapping")
		return make(map[int]map[string]string)
	}

	// Validate JSON format first
	jobMappingJSON = strings.TrimSpace(jobMappingJSON)
	if !strings.HasPrefix(jobMappingJSON, "{") || !strings.HasSuffix(jobMappingJSON, "}") {
		logger.Errorf("OLAKE_JOB_MAPPING must be a valid JSON object, got: %s", 
			truncateForLog(jobMappingJSON, 100))
		return make(map[int]map[string]string)
	}

	var jobMapping map[string]map[string]string
	if err := json.Unmarshal([]byte(jobMappingJSON), &jobMapping); err != nil {
		logger.Errorf("Failed to parse OLAKE_JOB_MAPPING as JSON: %v", err)
		logger.Errorf("Raw configuration (first 200 chars): %s", 
			truncateForLog(jobMappingJSON, 200))
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
		}

		// Validate node label keys and values
		validMapping := make(map[string]string)
		for key, value := range nodeLabels {
			if strings.TrimSpace(key) == "" {
				stats.InvalidMappings = append(stats.InvalidMappings, 
					fmt.Sprintf("JobID %d: empty label key", jobID))
				logger.Errorf("JobID %d has empty node label key", jobID)
				continue
			}
			if strings.TrimSpace(value) == "" {
				stats.InvalidMappings = append(stats.InvalidMappings, 
					fmt.Sprintf("JobID %d: empty label value for key '%s'", jobID, key))
				logger.Errorf("JobID %d has empty value for node label key '%s'", jobID, key)
				continue
			}
			validMapping[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}

		if len(validMapping) > 0 || len(nodeLabels) == 0 {
			result[jobID] = validMapping
			stats.ValidEntries++
		}
	}

	// Log comprehensive statistics
	logger.Infof("Job mapping loaded: %d valid entries out of %d total", 
		stats.ValidEntries, stats.TotalEntries)
	
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

// truncateForLog safely truncates a string for logging to prevent log spam
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}