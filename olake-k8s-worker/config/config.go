package config

import (
	"fmt"
	"os"
	"time"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/utils"
)

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

	// Logging configuration
	Logging LoggingConfig `json:"logging"`
}

// TemporalConfig contains Temporal-related settings
type TemporalConfig struct {
	Address   string        `json:"address"`
	TaskQueue string        `json:"task_queue"`
	Timeout   time.Duration `json:"timeout"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	URL      string `json:"url"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	RunMode  string `json:"run_mode"`

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// KubernetesConfig contains K8s-related settings
type KubernetesConfig struct {
	Namespace        string            `json:"namespace"`
	ImageRegistry    string            `json:"image_registry"`
	ImagePullPolicy  string            `json:"image_pull_policy"`
	ServiceAccount   string            `json:"service_account"`
	JobTTL           int32             `json:"job_ttl_seconds"`
	DefaultResources ResourceLimits    `json:"default_resources"`
	JobTimeout       time.Duration     `json:"job_timeout"`
	CleanupPolicy    string            `json:"cleanup_policy"`
	Labels           map[string]string `json:"labels"`
}

// ResourceLimits defines CPU and memory limits for K8s jobs
type ResourceLimits struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

// WorkerConfig contains worker-specific settings
type WorkerConfig struct {
	MaxConcurrentActivities int           `json:"max_concurrent_activities"`
	MaxConcurrentWorkflows  int           `json:"max_concurrent_workflows"`
	HeartbeatInterval       time.Duration `json:"heartbeat_interval"`
	WorkerIdentity          string        `json:"worker_identity"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Structured bool   `json:"structured"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Temporal: TemporalConfig{
			Address:   utils.GetEnv("TEMPORAL_ADDRESS", "temporal.default.svc.cluster.local:7233"),
			TaskQueue: "OLAKE_K8S_TASK_QUEUE", // Hardcoded as per requirement
			Timeout:   parseDuration("TEMPORAL_TIMEOUT", "30s"),
		},
		Database: DatabaseConfig{
			URL:             utils.GetEnv("DATABASE_URL", ""),
			Host:            utils.GetEnv("DB_HOST", "postgres.default.svc.cluster.local"),
			Port:            utils.GetEnv("DB_PORT", "5432"),
			User:            utils.GetEnv("DB_USER", "olake"),
			Password:        utils.GetEnv("DB_PASSWORD", "password"),
			Database:        utils.GetEnv("DB_NAME", "olake"),
			SSLMode:         utils.GetEnv("DB_SSLMODE", "disable"),
			RunMode:         utils.GetEnv("RUN_MODE", "production"),
			MaxOpenConns:    utils.GetEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    utils.GetEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: parseDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		Kubernetes: KubernetesConfig{
			Namespace:       utils.GetEnv("WORKER_NAMESPACE", "default"),
			ImageRegistry:   utils.GetEnv("IMAGE_REGISTRY", "olakego"),
			ImagePullPolicy: utils.GetEnv("IMAGE_PULL_POLICY", "IfNotPresent"),
			ServiceAccount:  utils.GetEnv("SERVICE_ACCOUNT", "olake-worker"),
			JobTTL:          int32(utils.GetEnvInt("JOB_TTL_SECONDS", 3600)), // 1 hour
			JobTimeout:      parseDuration("JOB_TIMEOUT", "15m"),
			CleanupPolicy:   utils.GetEnv("CLEANUP_POLICY", "auto"),
			DefaultResources: ResourceLimits{
				CPURequest:    utils.GetEnv("DEFAULT_CPU_REQUEST", "100m"),
				CPULimit:      utils.GetEnv("DEFAULT_CPU_LIMIT", "500m"),
				MemoryRequest: utils.GetEnv("DEFAULT_MEMORY_REQUEST", "256Mi"),
				MemoryLimit:   utils.GetEnv("DEFAULT_MEMORY_LIMIT", "1Gi"),
			},
			Labels: map[string]string{
				"app":        "olake-sync",
				"managed-by": "olake-k8s-worker",
				"version":    utils.GetEnv("WORKER_VERSION", "latest"),
			},
		},
		Worker: WorkerConfig{
			MaxConcurrentActivities: utils.GetEnvInt("MAX_CONCURRENT_ACTIVITIES", 10),
			MaxConcurrentWorkflows:  utils.GetEnvInt("MAX_CONCURRENT_WORKFLOWS", 5),
			HeartbeatInterval:       parseDuration("HEARTBEAT_INTERVAL", "5s"),
			WorkerIdentity:          generateWorkerIdentity(),
		},
		Logging: LoggingConfig{
			Level:      utils.GetEnv("LOG_LEVEL", "info"),
			Format:     utils.GetEnv("LOG_FORMAT", "console"),
			Structured: utils.GetEnvBool("LOG_STRUCTURED", false),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	logger.Info("Configuration loaded successfully")
	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate Temporal config
	if c.Temporal.Address == "" {
		return fmt.Errorf("temporal address is required")
	}
	if c.Temporal.TaskQueue == "" {
		return fmt.Errorf("temporal task queue is required")
	}

	// Validate Database config
	if c.Database.URL == "" {
		if c.Database.Host == "" || c.Database.User == "" || c.Database.Database == "" {
			return fmt.Errorf("database connection details are incomplete")
		}
	}

	// Validate Kubernetes config
	if c.Kubernetes.Namespace == "" {
		return fmt.Errorf("kubernetes namespace is required")
	}
	if c.Kubernetes.ImageRegistry == "" {
		return fmt.Errorf("image registry is required")
	}

	// Validate resource limits
	if err := c.validateResourceLimits(); err != nil {
		return fmt.Errorf("invalid resource limits: %w", err)
	}

	// Validate Worker config
	if c.Worker.MaxConcurrentActivities <= 0 {
		return fmt.Errorf("max concurrent activities must be positive")
	}
	if c.Worker.MaxConcurrentWorkflows <= 0 {
		return fmt.Errorf("max concurrent workflows must be positive")
	}

	return nil
}

// validateResourceLimits validates Kubernetes resource specifications
func (c *Config) validateResourceLimits() error {
	resources := c.Kubernetes.DefaultResources

	// Basic validation - ensure values are not empty
	if resources.CPURequest == "" || resources.CPULimit == "" {
		return fmt.Errorf("CPU request and limit must be specified")
	}
	if resources.MemoryRequest == "" || resources.MemoryLimit == "" {
		return fmt.Errorf("memory request and limit must be specified")
	}

	return nil
}

// GetDatabaseURL constructs the database URL from individual components
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetImageName constructs the full image name for a connector
func (c *Config) GetImageName(connectorType, version string) string {
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("%s/source-%s:%s", c.Kubernetes.ImageRegistry, connectorType, version)
}

// GetJobLabels returns default labels for Kubernetes jobs
func (c *Config) GetJobLabels() map[string]string {
	labels := make(map[string]string)
	for k, v := range c.Kubernetes.Labels {
		labels[k] = v
	}
	return labels
}

// Helper functions

func parseDuration(envKey, defaultValue string) time.Duration {
	value := utils.GetEnv(envKey, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		logger.Warnf("Invalid duration for %s: %s, using default: %s", envKey, value, defaultValue)
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

func generateWorkerIdentity() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	pid := os.Getpid()
	return fmt.Sprintf("olake-k8s-worker-%s-%d", hostname, pid)
}
