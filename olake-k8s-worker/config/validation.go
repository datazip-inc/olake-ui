package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ConfigValidator provides validation utilities for configuration
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
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
