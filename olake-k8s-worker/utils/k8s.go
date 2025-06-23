package utils

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// ParseQuantity parses Kubernetes resource quantity string
func ParseQuantity(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q
}

// ParseQuantityWithDefault parses quantity with fallback
func ParseQuantityWithDefault(s, defaultValue string) resource.Quantity {
	if s == "" {
		s = defaultValue
	}
	q, err := resource.ParseQuantity(s)
	if err != nil {
		q, _ = resource.ParseQuantity(defaultValue)
	}
	return q
}
