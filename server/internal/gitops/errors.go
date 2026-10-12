package gitops

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// ErrNonRetryable marks a sync error that will not succeed on retry
// (bad spec, invalid config, failed credential test).
var ErrNonRetryable = errors.New("non-retryable sync error")

// ErrStreamsNotFound means no Streams ConfigMap references this job yet.
var ErrStreamsNotFound = errors.New("streams resource not found")

func NonRetryableError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNonRetryable) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrNonRetryable, err)
}

// workflowError classifies an error from an OLake command run through Temporal: transient when the
// workflow never started (Temporal unreachable), non-retryable when it ran and failed.
func workflowError(err error) error {
	if err == nil || errors.Is(err, temporal.ErrWorkflowStart) {
		return err
	}
	return NonRetryableError(err)
}

func requireSpec(projectID, userID string) (int, error) {
	if projectID == "" {
		return 0, NonRetryableError(fmt.Errorf("data.projectId is required"))
	}
	id, err := strconv.Atoi(userID)
	if err != nil || id <= 0 {
		return 0, NonRetryableError(fmt.Errorf("data.userId must be a positive integer"))
	}
	return id, nil
}
