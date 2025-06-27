package types

import "time"

// WorkerConfig contains worker-specific settings
type WorkerConfig struct {
	MaxConcurrentActivities int           `json:"max_concurrent_activities"`
	MaxConcurrentWorkflows  int           `json:"max_concurrent_workflows"`
	HeartbeatInterval       time.Duration `json:"heartbeat_interval"`
	WorkerIdentity          string        `json:"worker_identity"`
}