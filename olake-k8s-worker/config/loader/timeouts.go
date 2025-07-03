package loader

import (
	"olake-k8s-worker/config/types"
	"olake-k8s-worker/utils/parser"
)

// LoadTimeouts loads timeout configuration from environment variables
func LoadTimeouts() (types.TimeoutConfig, error) {
	return types.TimeoutConfig{
		WorkflowExecution: types.WorkflowTimeouts{
			Discover: parser.ParseDuration("WORKFLOW_TIMEOUT_DISCOVER", "2h"), // 2 hours for discovery workflows
			Test:     parser.ParseDuration("WORKFLOW_TIMEOUT_TEST", "2h"),     // 2 hours for test workflows
			Sync:     parser.ParseDuration("WORKFLOW_TIMEOUT_SYNC", "720h"),   // 30 days for sync workflows
		},
		Activity: types.ActivityTimeouts{
			Discover: parser.ParseDuration("ACTIVITY_TIMEOUT_DISCOVER", "30m"), // 30 minutes for discovery activities
			Test:     parser.ParseDuration("ACTIVITY_TIMEOUT_TEST", "30m"),     // 30 minutes for test activities
			Sync:     parser.ParseDuration("ACTIVITY_TIMEOUT_SYNC", "700h"),    // 29 days for sync activities (effectively infinite)
		},
	}, nil
}