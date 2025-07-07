package loader

import (
	"olake-ui/olake-workers/k8s/config/types"
	"olake-ui/olake-workers/k8s/utils/parser"
)

// LoadTimeouts loads timeout configuration from environment variables
func LoadTimeouts() (types.TimeoutConfig, error) {
	return types.TimeoutConfig{
		WorkflowExecution: types.WorkflowTimeouts{
			Discover: parser.ParseDuration("TIMEOUT_WORKFLOW_DISCOVER", "3h"),   // 3 hours for discovery workflows
			Test:     parser.ParseDuration("TIMEOUT_WORKFLOW_TEST", "3h"),       // 3 hours for test workflows
			Sync:     parser.ParseDuration("TIMEOUT_WORKFLOW_SYNC", "720h"),     // 30 days for sync workflows
		},
		Activity: types.ActivityTimeouts{
			Discover: parser.ParseDuration("TIMEOUT_ACTIVITY_DISCOVER", "2h"),   // 2 hours for discovery activities
			Test:     parser.ParseDuration("TIMEOUT_ACTIVITY_TEST", "2h"),       // 2 hours for test activities
			Sync:     parser.ParseDuration("TIMEOUT_ACTIVITY_SYNC", "700h"),     // 29 days for sync activities
		},
	}, nil
}