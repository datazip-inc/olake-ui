package loader

import (
	"olake-ui/olake-workers/k8s/config/types"
	"olake-ui/olake-workers/k8s/utils/env"
	"olake-ui/olake-workers/k8s/utils/parser"
)

// LoadTemporal loads Temporal configuration from environment variables
func LoadTemporal() (types.TemporalConfig, error) {
	return types.TemporalConfig{
		Address:   env.GetEnv("TEMPORAL_ADDRESS", "temporal.olake.svc.cluster.local:7233"),
		TaskQueue: "OLAKE_K8S_TASK_QUEUE",
		Timeout:   parser.ParseDuration("TEMPORAL_TIMEOUT", "30s"),
	}, nil
}
