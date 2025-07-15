package k8s

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	
	"olake-ui/olake-workers/k8s/utils/env"
)

// ParseQuantity parses Kubernetes resource quantity string
func ParseQuantity(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q
}

// SanitizeName converts a string to a valid Kubernetes resource name
// Consolidates SanitizeK8sName and SanitizeKubernetesName functions
func SanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, ":", "-")

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Truncate if too long (max 63 characters for Kubernetes)
	if len(name) > 63 {
		name = name[:63]
		name = strings.TrimSuffix(name, "-")
	}

	return name
}

// GetCurrentPodName retrieves the current pod name from environment variable
func GetCurrentPodName() string {
	podName := env.GetEnv("POD_NAME", "")
	if podName == "" {
		panic("POD_NAME environment variable not set - check Kubernetes manifest")
	}
	return podName
}
