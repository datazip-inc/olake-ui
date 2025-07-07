package types

import (
	corev1 "k8s.io/api/core/v1"
)

// KubernetesConfig contains K8s-related settings
type KubernetesConfig struct {
	Namespace       string            `json:"namespace"`
	ImageRegistry   string            `json:"image_registry"`
	ImagePullPolicy string            `json:"image_pull_policy"`
	ServiceAccount  string            `json:"service_account"`
	PVCName         string            `json:"storage_pvc_name"`
	Labels          map[string]string `json:"labels"`
	JobScheduling   JobSchedulingConfig `json:"job_scheduling"`
}

// KubernetesResourceLimits defines CPU and memory limits for K8s jobs
type KubernetesResourceLimits struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

// JobSchedulingConfig contains scheduling configuration for different activity types
type JobSchedulingConfig struct {
	SyncJobs     ActivitySchedulingConfig `json:"sync"`
	DiscoverJobs ActivitySchedulingConfig `json:"discover"`
	TestJobs     ActivitySchedulingConfig `json:"test"`
}

// ActivitySchedulingConfig defines scheduling rules for a specific activity type
type ActivitySchedulingConfig struct {
	NodeSelector    map[string]string  `json:"node_selector"`
	Tolerations     []corev1.Toleration `json:"tolerations"`
	Affinity        *corev1.Affinity   `json:"affinity"`
	AntiAffinity    AntiAffinityConfig `json:"anti_affinity"`
}

// AntiAffinityConfig defines anti-affinity rules
type AntiAffinityConfig struct {
	Enabled      bool   `json:"enabled"`
	Strategy     string `json:"strategy"`     // "hard" or "soft"
	TopologyKey  string `json:"topology_key"` // default: "kubernetes.io/hostname"
	Weight       int32  `json:"weight"`       // used for soft anti-affinity (1-100)
}

