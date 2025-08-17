package config

import "time"

// Config holds all configuration for the K8s worker
type Config struct {
	// Temporal configuration
	Temporal TemporalConfig `mapstructure:"temporal"`

	// Database configuration
	Database DatabaseConfig `mapstructure:"database"`

	// Kubernetes configuration
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`

	// Worker configuration
	Worker WorkerConfig `mapstructure:"worker"`

	// Timeout configuration
	Timeouts TimeoutConfig `mapstructure:"timeouts"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`
}

// TemporalConfig contains Temporal-related settings
type TemporalConfig struct {
	Address   string        `mapstructure:"address"`
	TaskQueue string        `mapstructure:"task_queue"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	RunMode  string `mapstructure:"run_mode"`

	// Connection pool settings
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// KubernetesConfig contains K8s-related settings
type KubernetesConfig struct {
	Namespace         string                    `mapstructure:"namespace"`
	ImageRegistry     string                    `mapstructure:"image_registry"`
	ImagePullPolicy   string                    `mapstructure:"image_pull_policy"`
	ServiceAccount    string                    `mapstructure:"service_account"`
	PVCName           string                    `mapstructure:"storage_pvc_name"`
	Labels            map[string]string         `mapstructure:"labels"`
	JobMapping        map[int]map[string]string `mapstructure:"job_mapping"`
	JobServiceAccount string                    `mapstructure:"job_service_account"`
	OLakeSecretKey    string                    `mapstructure:"secret_key"`
}

// KubernetesResourceLimits defines CPU and memory limits for K8s jobs
type KubernetesResourceLimits struct {
	CPURequest    string `mapstructure:"cpu_request"`
	CPULimit      string `mapstructure:"cpu_limit"`
	MemoryRequest string `mapstructure:"memory_request"`
	MemoryLimit   string `mapstructure:"memory_limit"`
}

// WorkerConfig contains worker-specific settings
type WorkerConfig struct {
	MaxConcurrentActivities int           `mapstructure:"max_concurrent_activities"`
	MaxConcurrentWorkflows  int           `mapstructure:"max_concurrent_workflows"`
	HeartbeatInterval       time.Duration `mapstructure:"heartbeat_interval"`
	WorkerIdentity          string        `mapstructure:"worker_identity"`
}

// TimeoutConfig contains all timeout-related settings
type TimeoutConfig struct {
	// Workflow execution timeouts (client-side)
	WorkflowExecution WorkflowTimeouts `mapstructure:"workflow_execution"`

	// Activity timeouts (workflow-side)
	Activity ActivityTimeouts `mapstructure:"activity"`
}

// WorkflowTimeouts contains workflow execution timeout settings
type WorkflowTimeouts struct {
	Discover time.Duration `mapstructure:"discover"`
	Test     time.Duration `mapstructure:"test"`
	Sync     time.Duration `mapstructure:"sync"`
}

// ActivityTimeouts contains activity execution timeout settings
type ActivityTimeouts struct {
	Discover time.Duration `mapstructure:"discover"`
	Test     time.Duration `mapstructure:"test"`
	Sync     time.Duration `mapstructure:"sync"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}