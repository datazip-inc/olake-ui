package gitops

import (
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models"
)

type streamsShape struct {
	Streams         []json.RawMessage          `json:"streams"`
	SelectedStreams map[string]json.RawMessage `json:"selected_streams"`
}

// classifyStreamsFormat reports whether the Streams ConfigMap uses the split format
// (selected_streams only). Legacy format includes streams[] metadata in the CM.
// true: split format, false: legacy format.
func classifyStreamsFormat(config string) (split bool, err error) {
	var shape streamsShape
	if err := json.Unmarshal([]byte(config), &shape); err != nil {
		return false, NonRetryableError(fmt.Errorf("invalid streams config JSON: %w", err))
	}
	if len(shape.Streams) > 0 {
		return false, nil
	}
	if len(shape.SelectedStreams) > 0 {
		return true, nil
	}
	return false, NonRetryableError(fmt.Errorf("streams config is empty or has no selected_streams"))
}

func discoverJobID(job *models.Job) int {
	if job == nil {
		return -1
	}
	return job.ID
}

// streamsCMApplied is true when this Streams object was last synced as Ready for the current CM fingerprint
func streamsCMApplied(annotations map[string]string, data map[string]string) bool {
	if annotations == nil {
		return false
	}
	return annotations[AnnotationPhase] == PhaseReady &&
		annotations[AnnotationObservedHash] == ContentHash(data)
}
