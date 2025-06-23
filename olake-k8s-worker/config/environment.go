package config

import (
	"fmt"
	"strings"
)

// Environment represents different deployment environments
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// EnvironmentConfig provides environment-specific configurations
type EnvironmentConfig struct {
	Environment Environment `json:"environment"`

	// Environment-specific overrides
	LogLevel         string            `json:"log_level"`
	DatabaseSSLMode  string            `json:"database_ssl_mode"`
	JobTimeout       string            `json:"job_timeout"`
	CleanupPolicy    string            `json:"cleanup_policy"`
	ResourceLimits   ResourceLimits    `json:"resource_limits"`
	AdditionalLabels map[string]string `json:"additional_labels"`
}

// GetEnvironmentConfig returns configuration overrides for the current environment
func GetEnvironmentConfig(env string) *EnvironmentConfig {
	environment := Environment(strings.ToLower(env))

	switch environment {
	case Development:
		return &EnvironmentConfig{
			Environment:     Development,
			LogLevel:        "debug",
			DatabaseSSLMode: "disable",
			JobTimeout:      "30m",    // Longer timeout for development
			CleanupPolicy:   "manual", // Keep resources for debugging
			ResourceLimits: ResourceLimits{
				CPURequest:    "50m",
				CPULimit:      "200m",
				MemoryRequest: "128Mi",
				MemoryLimit:   "512Mi",
			},
			AdditionalLabels: map[string]string{
				"environment": "development",
				"debug":       "enabled",
			},
		}

	case Staging:
		return &EnvironmentConfig{
			Environment:     Staging,
			LogLevel:        "info",
			DatabaseSSLMode: "require",
			JobTimeout:      "20m",
			CleanupPolicy:   "auto",
			ResourceLimits: ResourceLimits{
				CPURequest:    "100m",
				CPULimit:      "300m",
				MemoryRequest: "256Mi",
				MemoryLimit:   "768Mi",
			},
			AdditionalLabels: map[string]string{
				"environment": "staging",
			},
		}

	case Production:
		return &EnvironmentConfig{
			Environment:     Production,
			LogLevel:        "info",
			DatabaseSSLMode: "require",
			JobTimeout:      "15m",
			CleanupPolicy:   "auto",
			ResourceLimits: ResourceLimits{
				CPURequest:    "100m",
				CPULimit:      "500m",
				MemoryRequest: "256Mi",
				MemoryLimit:   "1Gi",
			},
			AdditionalLabels: map[string]string{
				"environment": "production",
			},
		}

	default:
		// Return production config as safe default
		return GetEnvironmentConfig("production")
	}
}

// ApplyEnvironmentOverrides applies environment-specific overrides to the base config
func (c *Config) ApplyEnvironmentOverrides() {
	envConfig := GetEnvironmentConfig(c.Database.RunMode)

	// Apply overrides
	if envConfig.LogLevel != "" {
		c.Logging.Level = envConfig.LogLevel
	}

	if envConfig.DatabaseSSLMode != "" {
		c.Database.SSLMode = envConfig.DatabaseSSLMode
	}

	if envConfig.JobTimeout != "" {
		c.Kubernetes.JobTimeout = parseDuration("", envConfig.JobTimeout)
	}

	if envConfig.CleanupPolicy != "" {
		c.Kubernetes.CleanupPolicy = envConfig.CleanupPolicy
	}

	// Apply resource limits
	c.Kubernetes.DefaultResources = envConfig.ResourceLimits

	// Merge additional labels
	for k, v := range envConfig.AdditionalLabels {
		c.Kubernetes.Labels[k] = v
	}

	fmt.Printf("Applied %s environment configuration\n", envConfig.Environment)
}
