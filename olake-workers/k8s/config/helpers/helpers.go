package helpers

import (
	"fmt"
	"time"

	"olake-ui/olake-workers/k8s/config/types"
)

// GetWorkflowTimeout returns the workflow execution timeout for the given operation
func GetWorkflowTimeout(cfg *types.Config, operation string) time.Duration {
	switch operation {
	case "discover":
		return cfg.Timeouts.WorkflowExecution.Discover
	case "test":
		return cfg.Timeouts.WorkflowExecution.Test
	case "sync":
		return cfg.Timeouts.WorkflowExecution.Sync
	default:
		return time.Hour * 2 // Safe default
	}
}

// GetActivityTimeout returns the activity timeout for the given operation
func GetActivityTimeout(cfg *types.Config, operation string) time.Duration {
	switch operation {
	case "discover":
		return cfg.Timeouts.Activity.Discover
	case "test":
		return cfg.Timeouts.Activity.Test
	case "sync":
		return cfg.Timeouts.Activity.Sync
	default:
		return time.Minute * 30 // Safe default
	}
}

// GetDatabaseURL returns the database connection URL
func GetDatabaseURL(cfg *types.Config) string {
	if cfg.Database.URL != "" {
		return cfg.Database.URL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)
}

// GetImageName returns the full Docker image name for a connector
func GetImageName(cfg *types.Config, connectorType, version string) string {
	return fmt.Sprintf("%s/%s:%s", cfg.Kubernetes.ImageRegistry, connectorType, version)
}

// GetJobLabels returns a copy of the default labels for K8s jobs
func GetJobLabels(cfg *types.Config) map[string]string {
	labels := make(map[string]string)
	for k, v := range cfg.Kubernetes.Labels {
		labels[k] = v
	}
	return labels
}