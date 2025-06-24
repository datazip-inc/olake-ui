package utils

import (
	"encoding/json"
	"os"
	"strings"
)

// CreateDirectory creates a directory with the specified permissions if it doesn't exist
func CreateDirectory(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// WriteFile writes data to a file with the specified permissions
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// ParseJSONFile reads and parses a JSON file
func ParseJSONFile(filepath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SanitizeK8sName converts a string to a valid Kubernetes resource name
func SanitizeK8sName(name string) string {
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, ":", "-")
	return result
}
