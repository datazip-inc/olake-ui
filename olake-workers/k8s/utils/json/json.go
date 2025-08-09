package json

import (
	"encoding/json"
	"os"
)

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
