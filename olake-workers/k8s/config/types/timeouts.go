package types

import "time"

// TimeoutConfig contains all timeout-related settings
type TimeoutConfig struct {
	// Workflow execution timeouts (client-side)
	WorkflowExecution WorkflowTimeouts `json:"workflow_execution"`

	// Activity timeouts (workflow-side)
	Activity ActivityTimeouts `json:"activity"`
}

// WorkflowTimeouts contains workflow execution timeout settings
type WorkflowTimeouts struct {
	Discover time.Duration `json:"discover"`
	Test     time.Duration `json:"test"`
	Sync     time.Duration `json:"sync"`
}

// ActivityTimeouts contains activity execution timeout settings
type ActivityTimeouts struct {
	Discover time.Duration `json:"discover"`
	Test     time.Duration `json:"test"`
	Sync     time.Duration `json:"sync"`
}