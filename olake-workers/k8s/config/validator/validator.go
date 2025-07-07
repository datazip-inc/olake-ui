package validator

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"olake-ui/olake-workers/k8s/config/types"
)

// ConfigValidator provides validation utilities for configuration
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidateConfig validates the entire configuration
func (v *ConfigValidator) ValidateConfig(config *types.Config) error {
	// Validate Temporal config
	if err := v.ValidateTemporalConfig(&config.Temporal); err != nil {
		return fmt.Errorf("temporal configuration error: %w", err)
	}

	// Validate Database config
	if err := v.ValidateDatabaseConfig(&config.Database); err != nil {
		return fmt.Errorf("database configuration error: %w", err)
	}

	// Validate Kubernetes config
	if err := v.ValidateKubernetesConfig(&config.Kubernetes); err != nil {
		return fmt.Errorf("kubernetes configuration error: %w", err)
	}

	// Validate Worker config
	if err := v.ValidateWorkerConfig(&config.Worker); err != nil {
		return fmt.Errorf("worker configuration error: %w", err)
	}

	// Validate Timeout config
	if err := v.ValidateTimeoutConfig(&config.Timeouts); err != nil {
		return fmt.Errorf("timeout configuration error: %w", err)
	}

	return nil
}

// ValidateTemporalConfig validates Temporal configuration
func (v *ConfigValidator) ValidateTemporalConfig(config *types.TemporalConfig) error {
	if err := v.ValidateTemporalAddress(config.Address); err != nil {
		return err
	}

	if config.TaskQueue == "" {
		return fmt.Errorf("temporal task queue is required")
	}

	return nil
}

// ValidateDatabaseConfig validates database configuration
func (v *ConfigValidator) ValidateDatabaseConfig(config *types.DatabaseConfig) error {
	if config.URL != "" {
		return v.ValidateDatabaseURL(config.URL)
	}

	if config.Host == "" || config.User == "" || config.Database == "" {
		return fmt.Errorf("database connection details are incomplete")
	}

	return nil
}

// ValidateKubernetesConfig validates Kubernetes configuration
func (v *ConfigValidator) ValidateKubernetesConfig(config *types.KubernetesConfig) error {
	if err := v.ValidateKubernetesNamespace(config.Namespace); err != nil {
		return err
	}

	if err := v.ValidateImageRegistry(config.ImageRegistry); err != nil {
		return err
	}


	if err := v.ValidateLabels(config.Labels); err != nil {
		return err
	}

	return nil
}

// ValidateWorkerConfig validates worker configuration
func (v *ConfigValidator) ValidateWorkerConfig(config *types.WorkerConfig) error {
	if config.MaxConcurrentActivities <= 0 {
		return fmt.Errorf("max concurrent activities must be positive")
	}

	if config.MaxConcurrentWorkflows <= 0 {
		return fmt.Errorf("max concurrent workflows must be positive")
	}

	return nil
}

// ValidateTimeoutConfig validates timeout configuration
func (v *ConfigValidator) ValidateTimeoutConfig(config *types.TimeoutConfig) error {
	// Validate workflow timeouts
	if err := v.ValidateWorkflowTimeouts(&config.WorkflowExecution); err != nil {
		return fmt.Errorf("workflow timeout error: %w", err)
	}

	// Validate activity timeouts
	if err := v.ValidateActivityTimeouts(&config.Activity); err != nil {
		return fmt.Errorf("activity timeout error: %w", err)
	}

	// Validate that activity timeouts are less than workflow timeouts
	if err := v.ValidateTimeoutRelationships(config); err != nil {
		return fmt.Errorf("timeout relationship error: %w", err)
	}

	return nil
}

// ValidateWorkflowTimeouts validates workflow execution timeouts
func (v *ConfigValidator) ValidateWorkflowTimeouts(timeouts *types.WorkflowTimeouts) error {
	if timeouts.Discover <= 0 {
		return fmt.Errorf("workflow discover timeout must be positive")
	}
	if timeouts.Test <= 0 {
		return fmt.Errorf("workflow test timeout must be positive")
	}
	if timeouts.Sync <= 0 {
		return fmt.Errorf("workflow sync timeout must be positive")
	}

	// Check for reasonable minimums
	minTimeout := time.Minute * 5
	if timeouts.Discover < minTimeout {
		return fmt.Errorf("workflow discover timeout too short (minimum %v)", minTimeout)
	}
	if timeouts.Test < minTimeout {
		return fmt.Errorf("workflow test timeout too short (minimum %v)", minTimeout)
	}
	if timeouts.Sync < minTimeout {
		return fmt.Errorf("workflow sync timeout too short (minimum %v)", minTimeout)
	}

	return nil
}

// ValidateActivityTimeouts validates activity execution timeouts
func (v *ConfigValidator) ValidateActivityTimeouts(timeouts *types.ActivityTimeouts) error {
	if timeouts.Discover <= 0 {
		return fmt.Errorf("activity discover timeout must be positive")
	}
	if timeouts.Test <= 0 {
		return fmt.Errorf("activity test timeout must be positive")
	}
	if timeouts.Sync <= 0 {
		return fmt.Errorf("activity sync timeout must be positive")
	}

	// Check for reasonable minimums
	minTimeout := time.Minute * 1
	if timeouts.Discover < minTimeout {
		return fmt.Errorf("activity discover timeout too short (minimum %v)", minTimeout)
	}
	if timeouts.Test < minTimeout {
		return fmt.Errorf("activity test timeout too short (minimum %v)", minTimeout)
	}
	if timeouts.Sync < minTimeout {
		return fmt.Errorf("activity sync timeout too short (minimum %v)", minTimeout)
	}

	return nil
}

// ValidateTimeoutRelationships validates timeout relationships
func (v *ConfigValidator) ValidateTimeoutRelationships(config *types.TimeoutConfig) error {
	// Activity timeouts should be less than workflow timeouts
	if config.Activity.Discover >= config.WorkflowExecution.Discover {
		return fmt.Errorf("activity discover timeout (%v) must be less than workflow discover timeout (%v)",
			config.Activity.Discover, config.WorkflowExecution.Discover)
	}
	if config.Activity.Test >= config.WorkflowExecution.Test {
		return fmt.Errorf("activity test timeout (%v) must be less than workflow test timeout (%v)",
			config.Activity.Test, config.WorkflowExecution.Test)
	}
	if config.Activity.Sync >= config.WorkflowExecution.Sync {
		return fmt.Errorf("activity sync timeout (%v) must be less than workflow sync timeout (%v)",
			config.Activity.Sync, config.WorkflowExecution.Sync)
	}

	return nil
}

// ValidateResourceLimits validates Kubernetes resource specifications
func (v *ConfigValidator) ValidateResourceLimits(resources *types.KubernetesResourceLimits) error {
	// Basic validation - ensure values are not empty
	if resources.CPURequest == "" || resources.CPULimit == "" {
		return fmt.Errorf("CPU request and limit must be specified")
	}
	if resources.MemoryRequest == "" || resources.MemoryLimit == "" {
		return fmt.Errorf("memory request and limit must be specified")
	}

	// Validate each resource quantity format
	if err := v.ValidateResourceQuantity(resources.CPURequest); err != nil {
		return fmt.Errorf("invalid CPU request: %w", err)
	}
	if err := v.ValidateResourceQuantity(resources.CPULimit); err != nil {
		return fmt.Errorf("invalid CPU limit: %w", err)
	}
	if err := v.ValidateResourceQuantity(resources.MemoryRequest); err != nil {
		return fmt.Errorf("invalid memory request: %w", err)
	}
	if err := v.ValidateResourceQuantity(resources.MemoryLimit); err != nil {
		return fmt.Errorf("invalid memory limit: %w", err)
	}

	return nil
}

// ValidateTemporalAddress validates Temporal server address
func (v *ConfigValidator) ValidateTemporalAddress(address string) error {
	if address == "" {
		return fmt.Errorf("temporal address cannot be empty")
	}

	// Check if it's a valid host:port format
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid temporal address format: %w", err)
	}

	if host == "" {
		return fmt.Errorf("temporal host cannot be empty")
	}

	if portNum, err := strconv.Atoi(port); err != nil || portNum <= 0 || portNum > 65535 {
		return fmt.Errorf("invalid temporal port: %s", port)
	}

	return nil
}

// ValidateDatabaseURL validates database connection URL
func (v *ConfigValidator) ValidateDatabaseURL(dbURL string) error {
	if dbURL == "" {
		return fmt.Errorf("database URL cannot be empty")
	}

	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("invalid database URL format: %w", err)
	}

	if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		return fmt.Errorf("unsupported database scheme: %s", parsedURL.Scheme)
	}

	return nil
}

// ValidateKubernetesNamespace validates Kubernetes namespace name
func (v *ConfigValidator) ValidateKubernetesNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("kubernetes namespace cannot be empty")
	}

	// Kubernetes namespace naming rules
	matched, err := regexp.MatchString("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", namespace)
	if err != nil {
		return fmt.Errorf("error validating namespace format: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid kubernetes namespace format: %s", namespace)
	}

	if len(namespace) > 63 {
		return fmt.Errorf("kubernetes namespace too long (max 63 characters): %s", namespace)
	}

	return nil
}

// ValidateImageRegistry validates container image registry format
func (v *ConfigValidator) ValidateImageRegistry(registry string) error {
	if registry == "" {
		return fmt.Errorf("image registry cannot be empty")
	}

	// Basic validation for registry format
	if strings.Contains(registry, "://") {
		return fmt.Errorf("image registry should not include protocol: %s", registry)
	}

	return nil
}

// ValidateResourceQuantity validates Kubernetes resource quantity format
func (v *ConfigValidator) ValidateResourceQuantity(quantity string) error {
	if quantity == "" {
		return fmt.Errorf("resource quantity cannot be empty")
	}

	// Basic regex for Kubernetes quantity format (e.g., "100m", "1Gi", "500Mi")
	matched, err := regexp.MatchString("^[0-9]+(\\.[0-9]+)?(m|Mi|Gi|Ki|Ti)?$", quantity)
	if err != nil {
		return fmt.Errorf("error validating resource quantity format: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid resource quantity format: %s", quantity)
	}

	return nil
}

// ValidateLabels validates Kubernetes labels
func (v *ConfigValidator) ValidateLabels(labels map[string]string) error {
	for key, value := range labels {
		if err := v.validateLabelKey(key); err != nil {
			return fmt.Errorf("invalid label key %s: %w", key, err)
		}

		if err := v.validateLabelValue(value); err != nil {
			return fmt.Errorf("invalid label value %s: %w", value, err)
		}
	}

	return nil
}

func (v *ConfigValidator) validateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("label key cannot be empty")
	}

	if len(key) > 63 {
		return fmt.Errorf("label key too long (max 63 characters)")
	}

	matched, err := regexp.MatchString("^[a-z0-9A-Z]([a-z0-9A-Z._-]*[a-z0-9A-Z])?$", key)
	if err != nil {
		return fmt.Errorf("error validating label key format: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid label key format")
	}

	return nil
}

func (v *ConfigValidator) validateLabelValue(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("label value too long (max 63 characters)")
	}

	if value == "" {
		return nil // Empty values are allowed
	}

	matched, err := regexp.MatchString("^[a-z0-9A-Z]([a-z0-9A-Z._-]*[a-z0-9A-Z])?$", value)
	if err != nil {
		return fmt.Errorf("error validating label value format: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid label value format")
	}

	return nil
}