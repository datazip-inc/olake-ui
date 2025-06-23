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
