package types

import (
	"fmt"
	"time"
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

	// Timeout configuration
	Timeouts TimeoutConfig `json:"timeouts"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`
}

// GetWorkflowTimeout returns the workflow execution timeout for the given operation
func (c *Config) GetWorkflowTimeout(operation string) time.Duration {
	switch operation {
	case "discover":
		return c.Timeouts.WorkflowExecution.Discover
	case "test":
		return c.Timeouts.WorkflowExecution.Test
	case "sync":
		return c.Timeouts.WorkflowExecution.Sync
	default:
		return time.Hour * 2 // Safe default
	}
}

// GetActivityTimeout returns the activity timeout for the given operation
func (c *Config) GetActivityTimeout(operation string) time.Duration {
	switch operation {
	case "discover":
		return c.Timeouts.Activity.Discover
	case "test":
		return c.Timeouts.Activity.Test
	case "sync":
		return c.Timeouts.Activity.Sync
	default:
		return time.Minute * 30 // Safe default
	}
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetImageName returns the full Docker image name for a connector
func (c *Config) GetImageName(connectorType, version string) string {
	return fmt.Sprintf("%s/%s:%s", c.Kubernetes.ImageRegistry, connectorType, version)
}

// GetJobLabels returns a copy of the default labels for K8s jobs
func (c *Config) GetJobLabels() map[string]string {
	labels := make(map[string]string)
	for k, v := range c.Kubernetes.Labels {
		labels[k] = v
	}
	return labels
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Structured bool   `json:"structured"`
}

