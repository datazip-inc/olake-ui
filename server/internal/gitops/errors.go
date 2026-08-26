package gitops

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrPermanent marks a sync error that will not succeed on retry
// (bad spec, invalid config, failed credential test).
var ErrPermanent = errors.New("permanent sync error")

func Permanent(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrPermanent) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrPermanent, err)
}

func requireSpec(projectID, userID string) (int, error) {
	if projectID == "" {
		return 0, Permanent(fmt.Errorf("data.projectId is required"))
	}
	id, err := strconv.Atoi(userID)
	if err != nil || id <= 0 {
		return 0, Permanent(fmt.Errorf("data.userId must be a positive integer"))
	}
	return id, nil
}
