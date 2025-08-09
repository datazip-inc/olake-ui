package helpers

import (
	"time"

	"olake-ui/olake-workers/k8s/config/types"
)

// GetActivityTimeout returns the activity timeout for the given operation
func GetActivityTimeout(cfg *types.Config, operation string) time.Duration {
	switch operation {
	case "discover":
		return cfg.Timeouts.Activity.Discover
	case "test":
		return cfg.Timeouts.Activity.Test
	case "sync":
		return cfg.Timeouts.Activity.Sync
	default:
		return time.Minute * 30 // Safe default
	}
}
