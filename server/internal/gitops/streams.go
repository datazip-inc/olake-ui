package gitops

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

// streamsCM is a parsed Streams ConfigMap. Its shape picks the catalog format the job runs in:
// a self-contained CM (streams[] + selected_streams, the streams.json shape) runs in the legacy
// format; a selection-only CM (selected_streams alone) runs in the split format, with
// available_streams filled by discover.
type streamsCM struct {
	// Split is true for a selection-only CM.
	Split bool
	// Catalog is the CM's catalog: streams.json for a legacy CM, selected_streams.json for a
	// split one.
	Catalog string
}

func parseStreamsCM(config string) (streamsCM, error) {
	available, selected, err := utils.SplitCatalog(config)
	if err != nil {
		return streamsCM{}, NonRetryableError(fmt.Errorf("invalid streams config: %w", err))
	}
	if selected == "" {
		return streamsCM{}, NonRetryableError(fmt.Errorf("streams config has no selected_streams"))
	}
	if available == "" {
		return streamsCM{Split: true, Catalog: selected}, nil
	}
	return streamsCM{Catalog: config}, nil
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
