package gitops

import (
	"errors"
	"fmt"
)

// ErrPermanent marks a sync error that will not succeed on retry
// (bad spec, invalid config, failed credential test). Controllers
// should set phase=Failed and not requeue.
var ErrPermanent = errors.New("permanent sync error")

// Permanent wraps err so callers can detect it with errors.Is(err, ErrPermanent).
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrPermanent) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrPermanent, err)
}

func requireSpec(projectID string, userID int) error {
	if projectID == "" {
		return Permanent(fmt.Errorf("spec.projectId is required"))
	}
	if userID <= 0 {
		return Permanent(fmt.Errorf("spec.userId is required"))
	}
	return nil
}
