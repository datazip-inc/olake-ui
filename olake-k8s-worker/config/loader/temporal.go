package loader

import (
	"olake-k8s-worker/config/types"
	"olake-k8s-worker/utils/env"
	"olake-k8s-worker/utils/parser"
)

// LoadTemporal loads Temporal configuration from environment variables
func LoadTemporal() (types.TemporalConfig, error) {
	return types.TemporalConfig{
		Address:   env.GetEnv("TEMPORAL_ADDRESS", "temporal.olake.svc.cluster.local:7233"),
		TaskQueue: "OLAKE_K8S_TASK_QUEUE", // Hardcoded as per requirement
		Timeout:   parser.ParseDuration("TEMPORAL_TIMEOUT", "30s"),
	}, nil
}