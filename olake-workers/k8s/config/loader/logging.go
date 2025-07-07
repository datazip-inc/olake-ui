package loader

import (
	"olake-ui/olake-workers/k8s/config/types"
	"olake-ui/olake-workers/k8s/utils/env"
)

// LoadLogging loads logging configuration from environment variables
func LoadLogging() (types.LoggingConfig, error) {
	return types.LoggingConfig{
		Level:  env.GetEnv("LOG_LEVEL", "info"),
		Format: env.GetEnv("LOG_FORMAT", "console"),
	}, nil
}
