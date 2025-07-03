package config

import (
	"fmt"

	"olake-k8s-worker/config/loader"
	"olake-k8s-worker/config/types"
	"olake-k8s-worker/config/validator"
)

// Re-export types for backward compatibility
type Config = types.Config
type TemporalConfig = types.TemporalConfig
type DatabaseConfig = types.DatabaseConfig
type KubernetesConfig = types.KubernetesConfig
type WorkerConfig = types.WorkerConfig
type TimeoutConfig = types.TimeoutConfig
type WorkflowTimeouts = types.WorkflowTimeouts
type ActivityTimeouts = types.ActivityTimeouts
type ResourceLimits = types.KubernetesResourceLimits
type LoggingConfig = types.LoggingConfig

// Constants for backward compatibility

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load configuration directly
	config, err := loader.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration using the validator
	configValidator := validator.NewConfigValidator()
	if err := configValidator.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}


// Legacy functions for backward compatibility (deprecated)

// NewConfigValidator creates a new configuration validator
// Deprecated: Use validator.NewConfigValidator() directly
func NewConfigValidator() *validator.ConfigValidator {
	return validator.NewConfigValidator()
}

