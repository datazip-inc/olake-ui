package loader

import (
	"olake-k8s-worker/config/types"
	"olake-k8s-worker/utils/env"
)

// LoadLogging loads logging configuration from environment variables
func LoadLogging() (types.LoggingConfig, error) {
	return types.LoggingConfig{
		Level:      env.GetEnv("LOG_LEVEL", "info"),
		Format:     env.GetEnv("LOG_FORMAT", "console"),
		Structured: env.GetEnvBool("LOG_STRUCTURED", false),
	}, nil
}