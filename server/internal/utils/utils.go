package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ConnectionStatus represents the structure of the connection status JSON
type ConnectionStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// LogMessage represents the structure of the log message JSON
type LogMessage struct {
	ConnectionStatus *ConnectionStatus `json:"connectionStatus,omitempty"`
	Type             string            `json:"type,omitempty"`
	// Add other fields as needed
}

// ExtractAndParseLastLogMessage extracts the JSON from the last log line and parses it
func ExtractAndParseLastLogMessage(output []byte) (*LogMessage, error) {
	// Convert output to string and split into lines
	outputStr := strings.TrimSpace(string(output))
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

	// Parse the JSON
	var msg LogMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &msg, nil
}
