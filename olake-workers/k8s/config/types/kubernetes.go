package types

// KubernetesConfig contains K8s-related settings
type KubernetesConfig struct {
	Namespace         string                    `json:"namespace"`
	ImageRegistry     string                    `json:"image_registry"`
	ImagePullPolicy   string                    `json:"image_pull_policy"`
	ServiceAccount    string                    `json:"service_account"`
	PVCName           string                    `json:"storage_pvc_name"`
	Labels            map[string]string         `json:"labels"`
	JobMapping        map[int]map[string]string `json:"job_mapping"`
	JobServiceAccount string                    `json:"job_service_account"`
}

// KubernetesResourceLimits defines CPU and memory limits for K8s jobs
type KubernetesResourceLimits struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}
