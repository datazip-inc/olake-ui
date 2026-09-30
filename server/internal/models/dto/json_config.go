package dto

import (
	"encoding/json"
	"fmt"
)

// JSONConfig accepts config as either a JSON-encoded string or object on input,
// and always stores it as a compact JSON string internally.
type JSONConfig string

func (f *JSONConfig) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = JSONConfig(s)
		return nil
	}

	if len(data) > 0 && data[0] == '{' {
		var obj map[string]interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("config: invalid JSON object: %w", err)
		}
		compact, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("config: cannot re-marshal object: %w", err)
		}
		*f = JSONConfig(compact)
		return nil
	}

	return fmt.Errorf("config: must be a JSON string or object")
}

func (f JSONConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f JSONConfig) String() string {
	return string(f)
}
