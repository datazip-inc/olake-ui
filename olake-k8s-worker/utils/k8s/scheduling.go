package k8s

import (
	"fmt"

	"olake-k8s-worker/config/types"
	"olake-k8s-worker/utils/env"
	"olake-k8s-worker/utils/parser"
	corev1 "k8s.io/api/core/v1"
)

// GetDefaultJobSchedulingConfig returns default scheduling configuration for Kubernetes jobs
func GetDefaultJobSchedulingConfig() types.JobSchedulingConfig {
	return types.JobSchedulingConfig{
		SyncJobs: types.ActivitySchedulingConfig{
			NodeSelector: map[string]string{},
			Tolerations:  []corev1.Toleration{},
			Affinity:     nil,
			AntiAffinity: types.AntiAffinityConfig{
				Enabled:     true,
				Strategy:    "hard",
				TopologyKey: "kubernetes.io/hostname",
				Weight:      100,
			},
		},
		DiscoverJobs: types.ActivitySchedulingConfig{
			NodeSelector: map[string]string{},
			Tolerations:  []corev1.Toleration{},
			Affinity:     nil,
			AntiAffinity: types.AntiAffinityConfig{
				Enabled:     false,
				Strategy:    "soft",
				TopologyKey: "kubernetes.io/hostname",
				Weight:      50,
			},
		},
		TestJobs: types.ActivitySchedulingConfig{
			NodeSelector: map[string]string{},
			Tolerations:  []corev1.Toleration{},
			Affinity:     nil,
			AntiAffinity: types.AntiAffinityConfig{
				Enabled:     false,
				Strategy:    "soft",
				TopologyKey: "kubernetes.io/hostname",
				Weight:      50,
			},
		},
	}
}

// LoadJobSchedulingFromEnv loads job scheduling configuration from environment variables
// Following Airbyte pattern: OLAKE_{ACTIVITY}_JOB_{SETTING}
func LoadJobSchedulingFromEnv(config types.JobSchedulingConfig) types.JobSchedulingConfig {
	// Load sync job configuration
	config.SyncJobs = loadActivitySchedulingFromEnv("SYNC", config.SyncJobs)
	
	// Load discover job configuration
	config.DiscoverJobs = loadActivitySchedulingFromEnv("DISCOVER", config.DiscoverJobs)
	
	// Load test job configuration
	config.TestJobs = loadActivitySchedulingFromEnv("TEST", config.TestJobs)
	
	return config
}

// loadActivitySchedulingFromEnv loads activity-specific scheduling configuration
func loadActivitySchedulingFromEnv(activity string, config types.ActivitySchedulingConfig) types.ActivitySchedulingConfig {
	// Load node selector from environment variable
	// Format: OLAKE_SYNC_JOB_NODE_SELECTOR="key1=value1,key2=value2"
	nodeSelectorEnv := env.GetEnv(fmt.Sprintf("OLAKE_%s_JOB_NODE_SELECTOR", activity), "")
	if nodeSelectorEnv != "" {
		nodeSelector := parser.ParseKeyValuePairs(nodeSelectorEnv)
		if len(nodeSelector) > 0 {
			config.NodeSelector = nodeSelector
		}
	}
	
	// Load anti-affinity strategy
	// Format: OLAKE_SYNC_JOB_ANTI_AFFINITY_STRATEGY="hard"
	antiAffinityStrategy := env.GetEnv(fmt.Sprintf("OLAKE_%s_JOB_ANTI_AFFINITY_STRATEGY", activity), "")
	if antiAffinityStrategy != "" {
		config.AntiAffinity.Strategy = antiAffinityStrategy
	}
	
	// Load anti-affinity topology key
	// Format: OLAKE_SYNC_JOB_ANTI_AFFINITY_TOPOLOGY_KEY="kubernetes.io/hostname"
	topologyKey := env.GetEnv(fmt.Sprintf("OLAKE_%s_JOB_ANTI_AFFINITY_TOPOLOGY_KEY", activity), "")
	if topologyKey != "" {
		config.AntiAffinity.TopologyKey = topologyKey
	}
	
	// Load anti-affinity enabled flag
	// Format: OLAKE_SYNC_JOB_ANTI_AFFINITY_ENABLED="true"
	antiAffinityEnabledEnv := env.GetEnv(fmt.Sprintf("OLAKE_%s_JOB_ANTI_AFFINITY_ENABLED", activity), "")
	if antiAffinityEnabledEnv != "" {
		config.AntiAffinity.Enabled = env.GetEnvBool(fmt.Sprintf("OLAKE_%s_JOB_ANTI_AFFINITY_ENABLED", activity), config.AntiAffinity.Enabled)
	}
	
	// Load anti-affinity weight
	// Format: OLAKE_SYNC_JOB_ANTI_AFFINITY_WEIGHT="100"
	weightEnv := env.GetEnv(fmt.Sprintf("OLAKE_%s_JOB_ANTI_AFFINITY_WEIGHT", activity), "")
	if weightEnv != "" {
		weight := env.GetEnvInt(fmt.Sprintf("OLAKE_%s_JOB_ANTI_AFFINITY_WEIGHT", activity), int(config.AntiAffinity.Weight))
		config.AntiAffinity.Weight = int32(weight)
	}
	
	return config
}