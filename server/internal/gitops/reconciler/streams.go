package reconciler

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/types"
)

// parseStreamsCM returns the Streams ConfigMap's catalog: a selection-only CM in the
// selected_streams.json shape (Selected); discover fills the available streams. The legacy
// streams.json shape (streams[] + selected_streams) is rejected.
func parseStreamsCM(config string) (types.StreamsCatalog, error) {
	var cm struct {
		Streams         json.RawMessage `json:"streams"`
		SelectedStreams json.RawMessage `json:"selected_streams"`
	}
	if err := json.Unmarshal([]byte(config), &cm); err != nil {
		return types.StreamsCatalog{}, NonRetryableError(fmt.Errorf("invalid streams config: %w", err))
	}
	if !isEmptyJSON(cm.Streams) {
		return types.StreamsCatalog{}, NonRetryableError(fmt.Errorf("legacy streams config (streams[]) is not supported; keep only selected_streams (streams v2)"))
	}
	if isEmptyJSON(cm.SelectedStreams) {
		return types.StreamsCatalog{}, NonRetryableError(fmt.Errorf("streams config has no selected_streams"))
	}
	return types.StreamsCatalog{Selected: config}, nil
}

// isEmptyJSON reports whether a JSON value is missing, null, or an empty array or object.
func isEmptyJSON(value json.RawMessage) bool {
	switch string(bytes.TrimSpace(value)) {
	case "", "null", "[]", "{}":
		return true
	}
	return false
}

// discoverJobID is the JobID for a discover request: the existing job's ID, or -1 on create.
func discoverJobID(job *models.Job) int {
	if job == nil {
		return -1
	}
	return job.ID
}

// streamsCMApplied is true when this Streams object was last synced as Ready for the current CM fingerprint
func streamsCMApplied(annotations, data map[string]string) bool {
	if annotations == nil {
		return false
	}
	return annotations[AnnotationPhase] == PhaseReady &&
		annotations[AnnotationObservedHash] == ContentHash(data)
}
