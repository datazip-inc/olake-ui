package env

import (
	"os"
	"strconv"
)

// GetEnv returns environment variable value or default if not set
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt returns environment variable as int or default
func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}
	return defaultValue
}
