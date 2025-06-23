package database

import (
	"encoding/json"
	"fmt"
	"olake-k8s-worker/logger"
)

// Database operations for activities

// ParseJobOutput extracts meaningful information from job output logs
func ParseJobOutput(output string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Try to find JSON in the output (similar to Docker implementation)
	lines := splitLines(output)

	for _, line := range lines {
		// Try to parse each line as JSON
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
			// Found valid JSON, merge with result
			for k, v := range jsonData {
				result[k] = v
			}
		}
	}

	// If no JSON found, return basic info
	if len(result) == 0 {
		result["raw_output"] = output
		result["status"] = "completed"
	}

	return result, nil
}

// ValidateJobConfig checks if job configuration is valid
func ValidateJobConfig(jobData *JobData) error {
	if jobData.ID == 0 {
		return fmt.Errorf("invalid job ID")
	}

	if jobData.SourceType == "" {
		return fmt.Errorf("source type is required")
	}

	if jobData.DestType == "" {
		return fmt.Errorf("destination type is required")
	}

	if !jobData.Active {
		return fmt.Errorf("job is not active")
	}

	return nil
}

// LogJobExecution logs job execution details for monitoring
func LogJobExecution(jobData *JobData, command string, status string) {
	logger.Infof("Job Execution - ID: %d, Name: %s, Command: %s, Status: %s",
		jobData.ID, jobData.Name, command, status)
}

// Helper functions
func splitLines(text string) []string {
	lines := []string{}
	current := ""

	for _, char := range text {
		if char == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}
