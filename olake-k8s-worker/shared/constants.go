package shared

const (
	// Hardcoded - K8s worker only listens to this queue
	TaskQueue = "OLAKE_K8S_TASK_QUEUE"

	// Default FQDN for Temporal service in K8s cluster
	DefaultTemporalAddress = "temporal.default.svc.cluster.local:7233"
)
