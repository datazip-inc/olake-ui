package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/utils/env"
)

// SplitLines splits text into lines, removing empty lines
func SplitLines(text string) []string {
	lines := []string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// ParseJobOutput extracts meaningful information from job output logs
// This is a flexible parser that can handle different types of outputs
func ParseJobOutput(output string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	outputStr := strings.TrimSpace(output)

	if outputStr == "" {
		return nil, fmt.Errorf("empty output")
	}

	// Special handling for connection test outputs
	if strings.Contains(outputStr, "connectionStatus") {
		return parseConnectionTestOutput(outputStr)
	}

	// Try to find JSON in the output (similar to Docker implementation)
	lines := SplitLines(outputStr)

	for _, line := range lines {
		// Try to parse each line as JSON
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
			// Found valid JSON, merge with result
			for k, v := range jsonData {
				result[k] = v
			}
		} else {
			// Try to extract JSON part from the line (everything after the first "{")
			start := strings.Index(line, "{")
			if start != -1 {
				jsonStr := line[start:]
				if err := json.Unmarshal([]byte(jsonStr), &jsonData); err == nil {
					// Found valid JSON after prefix, merge with result
					for k, v := range jsonData {
						result[k] = v
					}
				}
			}
		}
	}

	// If no JSON found, return basic info
	if len(result) == 0 {
		result["raw_output"] = outputStr
		result["status"] = "completed"
	}

	return result, nil
}

// parseConnectionTestOutput is a private helper function for connection test parsing
// Matches the logic from server/utils/utils.go ExtractAndParseLastLogMessage
func parseConnectionTestOutput(output string) (map[string]interface{}, error) {
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

// ParseDuration parses a duration string with error handling and fallback
func ParseDuration(envKey, defaultValue string) time.Duration {
	value := env.GetEnv(envKey, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		logger.Warnf("Failed to parse duration for %s, using default: %s", envKey, defaultValue)
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

// GetOptionalTTL converts an environment variable to an optional TTL pointer
func GetOptionalTTL(envKey string, defaultValue int) *int32 {
	value := env.GetEnvInt(envKey, defaultValue)
	if value <= 0 {
		return nil
	}
	ttl := int32(value)
	return &ttl
}

// ParseKeyValuePairs parses a string of key=value pairs separated by commas
// Example: "key1=value1,key2=value2" -> map[string]string{"key1": "value1", "key2": "value2"}
func ParseKeyValuePairs(input string) map[string]string {
	result := make(map[string]string)
	if input == "" {
		return result
	}
	
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if key != "" {
				result[key] = value
			}
		}
	}
	
	return result
}