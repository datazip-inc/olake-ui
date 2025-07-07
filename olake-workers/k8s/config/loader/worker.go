package loader

import (
	"olake-ui/olake-workers/k8s/config/types"
	"olake-ui/olake-workers/k8s/utils/env"
	"olake-ui/olake-workers/k8s/utils/k8s"
	"olake-ui/olake-workers/k8s/utils/parser"
)

// LoadWorker loads worker configuration from environment variables
func LoadWorker() (types.WorkerConfig, error) {
	return types.WorkerConfig{
		MaxConcurrentActivities: env.GetEnvInt("MAX_CONCURRENT_ACTIVITIES", 10),
		MaxConcurrentWorkflows:  env.GetEnvInt("MAX_CONCURRENT_WORKFLOWS", 5),
		HeartbeatInterval:       parser.ParseDuration("HEARTBEAT_INTERVAL", "5s"),
		WorkerIdentity:          k8s.GenerateWorkerIdentity(),
	}, nil
}