package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSON converts any struct to JSON string
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON converts JSON string to struct
func FromJSON(jsonStr string, v interface{}) error {
	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		return fmt.Errorf("failed to unmarshal from JSON: %w", err)
	}
	return nil
}

// PrettyJSON returns formatted JSON string
func PrettyJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to pretty JSON: %w", err)
	}
	return string(data), nil
}
