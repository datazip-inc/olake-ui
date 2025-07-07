package types

// Config holds all configuration for the K8s worker
type Config struct {
	// Temporal configuration
	Temporal TemporalConfig `json:"temporal"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Kubernetes configuration
	Kubernetes KubernetesConfig `json:"kubernetes"`

	// Worker configuration
	Worker WorkerConfig `json:"worker"`

	// Timeout configuration
	Timeouts TimeoutConfig `json:"timeouts"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`
}
