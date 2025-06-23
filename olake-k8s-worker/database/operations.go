package database

import (
	"encoding/json"
	"fmt"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/utils"
	"strings"
)

// Database operations for activities

// ParseJobOutput extracts meaningful information from job output logs
func ParseJobOutput(output string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Try to find JSON in the output (similar to Docker implementation)
	lines := utils.SplitLines(output)

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

// ParseConnectionTestOutput specifically parses connection test results like Docker worker
func ParseConnectionTestOutput(output string) (map[string]interface{}, error) {
	// Use same logic as server/utils/utils.go ExtractAndParseLastLogMessage
	outputStr := strings.TrimSpace(output)
	if outputStr == "" {
		return nil, fmt.Errorf("empty output")
	}

	lines := strings.Split(outputStr, "\n")

	// Find the last non-empty line
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			lastLine = trimmed
			break
		}
	}

	if lastLine == "" {
		return nil, fmt.Errorf("no log lines found")
	}

	// Extract JSON part (everything after the first "{")
	start := strings.Index(lastLine, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON found in log line")
	}
	jsonStr := lastLine[start:]

	// Parse the JSON as LogMessage
	var logMessage struct {
		ConnectionStatus *struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"connectionStatus,omitempty"`
		Type string `json:"type,omitempty"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &logMessage); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if logMessage.ConnectionStatus == nil {
		return nil, fmt.Errorf("connection status not found")
	}

	return map[string]interface{}{
		"message": logMessage.ConnectionStatus.Message,
		"status":  logMessage.ConnectionStatus.Status,
	}, nil
}
