package shared

const (
	// Hardcoded - K8s worker only listens to this queue
	TaskQueue = "OLAKE_K8S_TASK_QUEUE"

	// Default FQDN for Temporal service in K8s cluster
	DefaultTemporalAddress = "temporal.default.svc.cluster.local:7233"
)

// Kubernetes constants
const (
	DefaultNamespace       = "default"
	DefaultImageRegistry   = "olakego"
	DefaultImagePullPolicy = "IfNotPresent"
	DefaultServiceAccount  = "olake-worker"
	DefaultJobTTLSeconds   = 86400 // 24 hours for long-running jobs
	DefaultCleanupPolicy   = "auto"

	// Resource defaults
	DefaultCPURequest    = "100m"
	DefaultCPULimit      = "500m"
	DefaultMemoryRequest = "256Mi"
	DefaultMemoryLimit   = "1Gi"
)

// Label constants (keep existing labels)
const (
	LabelApp        = "app"
	LabelType       = "type"
	LabelOperation  = "operation"
	LabelCleanup    = "cleanup"
	LabelManagedBy  = "managed-by"
	LabelVersion    = "version"
	LabelJobID      = "job-id"
	LabelWorkflowID = "workflow-id"

	// Label values
	LabelValueOlakeSync    = "olake-sync"
	LabelValueConnectorJob = "connector-job"
	LabelValueJobConfig    = "job-config"
	LabelValueK8sWorker    = "olake-k8s-worker"
)
