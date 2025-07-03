package loader

import (
	"olake-k8s-worker/config/types"
	"olake-k8s-worker/utils/env"
	"olake-k8s-worker/utils/k8s"
	"olake-k8s-worker/utils/parser"
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