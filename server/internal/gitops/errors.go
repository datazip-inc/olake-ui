package gitops

import (
	"errors"
	"fmt"
	"strconv"
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
