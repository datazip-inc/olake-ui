package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

// ParseConnectionTestOutput specifically parses connection test results like Docker worker
// Matches the logic from server/utils/utils.go ExtractAndParseLastLogMessage
func ParseConnectionTestOutput(output string) (map[string]interface{}, error) {
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

// GenerateJobName creates a Kubernetes-compatible job name
func GenerateJobName(prefix string, identifier string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, identifier, time.Now().Unix())
}

// GenerateWorkflowID generates a unique workflow ID
func GenerateWorkflowID(prefix string, jobID int) string {
	return fmt.Sprintf("%s-%d-%d", prefix, jobID, time.Now().Unix())
}

// SanitizeKubernetesName ensures the name follows Kubernetes naming conventions
func SanitizeKubernetesName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Truncate if too long (max 63 characters for Kubernetes)
	if len(name) > 63 {
		name = name[:63]
		name = strings.TrimSuffix(name, "-")
	}

	return name
}
