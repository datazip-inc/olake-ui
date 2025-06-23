package config

import "time"

// Global timeout configuration (loaded once at startup)
var GlobalTimeouts *TimeoutConfig

// InitGlobalTimeouts initializes the global timeout configuration
func InitGlobalTimeouts(config *Config) {
	GlobalTimeouts = &config.Timeouts
}

// GetWorkflowTimeout returns the workflow execution timeout for the given operation
func GetWorkflowTimeout(operation string) time.Duration {
	if GlobalTimeouts == nil {
		// Fallback to safe defaults if config not loaded
		return time.Hour * 2
	}

	switch operation {
	case "discover":
		return GlobalTimeouts.WorkflowExecution.Discover
	case "test":
		return GlobalTimeouts.WorkflowExecution.Test
	case "sync":
		return GlobalTimeouts.WorkflowExecution.Sync
	default:
		return time.Hour * 2 // Safe default
	}
}

// GetActivityTimeout returns the activity timeout for the given operation
func GetActivityTimeout(operation string) time.Duration {
	if GlobalTimeouts == nil {
		// Fallback to safe defaults if config not loaded
		return time.Minute * 30
	}

	switch operation {
	case "discover":
		return GlobalTimeouts.Activity.Discover
	case "test":
		return GlobalTimeouts.Activity.Test
	case "sync":
		return GlobalTimeouts.Activity.Sync
	default:
		return time.Minute * 30 // Safe default
	}
}

// DefaultTimeouts returns reasonable default timeout values
func DefaultTimeouts() TimeoutConfig {
	return TimeoutConfig{
		WorkflowExecution: WorkflowTimeouts{
			Discover: time.Hour * 2,
			Test:     time.Hour * 2,
			Sync:     time.Hour * 6,
		},
		Activity: ActivityTimeouts{
			Discover: time.Minute * 30,
			Test:     time.Minute * 30,
			Sync:     time.Hour * 4,
		},
	}
}
