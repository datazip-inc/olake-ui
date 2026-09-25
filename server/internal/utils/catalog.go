package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// SplitCatalog splits a combined {"streams": [...], "selected_streams": {...}} catalog into the
// contents of available_streams.json and selected_streams.json
func SplitCatalog(combinedCatalog string) (available, selected string, err error) {
	var catalog struct {
		Streams         json.RawMessage `json:"streams"`
		SelectedStreams json.RawMessage `json:"selected_streams"`
	}
	if err := json.Unmarshal([]byte(combinedCatalog), &catalog); err != nil {
		return "", "", fmt.Errorf("invalid catalog JSON: %w", err)
	}

	if available, err = wrapCatalogKey("streams", catalog.Streams); err != nil {
		return "", "", err
	}
	if selected, err = wrapCatalogKey("selected_streams", catalog.SelectedStreams); err != nil {
		return "", "", err
	}
	return available, selected, nil
}

func wrapCatalogKey(key string, value json.RawMessage) (string, error) {
	trimmed := bytes.TrimSpace(value)
	switch string(trimmed) {
	case "", "null", "[]", "{}":
		return "", nil
	}
	out, err := json.Marshal(map[string]json.RawMessage{key: trimmed})
	if err != nil {
		return "", err
	}
	return string(out), nil
}
