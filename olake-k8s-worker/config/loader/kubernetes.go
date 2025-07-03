package loader

import (
	"olake-k8s-worker/config/types"
	"olake-k8s-worker/utils/env"
	"olake-k8s-worker/utils/k8s"
	"olake-k8s-worker/utils/parser"
)

// LoadKubernetes loads Kubernetes configuration from environment variables
func LoadKubernetes() (types.KubernetesConfig, error) {
	// Load default job scheduling configuration
	jobScheduling := k8s.GetDefaultJobSchedulingConfig()
	
	// Override with environment variables if provided
	jobScheduling = k8s.LoadJobSchedulingFromEnv(jobScheduling)
	
	return types.KubernetesConfig{
		Namespace:       env.GetEnv("WORKER_NAMESPACE", "olake"),
		ImageRegistry:   env.GetEnv("IMAGE_REGISTRY", "olakego"),
		ImagePullPolicy: env.GetEnv("IMAGE_PULL_POLICY", "IfNotPresent"),
		ServiceAccount:  env.GetEnv("SERVICE_ACCOUNT", "olake-worker"),
		PVCName:         env.GetEnv("OLAKE_STORAGE_PVC_NAME", "olake-jobs-pvc"),
		JobTTL:          parser.GetOptionalTTL("JOB_TTL_SECONDS", 0),
		JobTimeout:      parser.ParseDuration("JOB_TIMEOUT", "15m"),
		CleanupPolicy:   env.GetEnv("CLEANUP_POLICY", "auto"),
		Labels: map[string]string{
			"app":        "olake-sync",
			"managed-by": "olake-k8s-worker",
			"version":    env.GetEnv("WORKER_VERSION", "latest"),
		},
		JobScheduling:   jobScheduling,
	}, nil
}