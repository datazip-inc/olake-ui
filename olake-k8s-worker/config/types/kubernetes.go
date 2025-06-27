package types

import "time"

// KubernetesConfig contains K8s-related settings
type KubernetesConfig struct {
	Namespace        string                   `json:"namespace"`
	ImageRegistry    string                   `json:"image_registry"`
	ImagePullPolicy  string                   `json:"image_pull_policy"`
	ServiceAccount   string                   `json:"service_account"`
	JobTTL           *int32                   `json:"job_ttl_seconds"`
	JobTimeout       time.Duration            `json:"job_timeout"`
	CleanupPolicy    string                   `json:"cleanup_policy"`
	Labels           map[string]string        `json:"labels"`
}

// KubernetesResourceLimits defines CPU and memory limits for K8s jobs
type KubernetesResourceLimits struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}