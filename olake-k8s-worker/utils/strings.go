package utils

import (
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
