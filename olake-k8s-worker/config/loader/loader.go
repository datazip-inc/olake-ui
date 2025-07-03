package loader

import (
	"fmt"

	"olake-k8s-worker/config/types"
	"olake-k8s-worker/logger"
)

// LoadConfig loads the complete configuration from environment variables
func LoadConfig() (*types.Config, error) {
	config := &types.Config{}

	// Load each configuration section
	var err error
	
	config.Temporal, err = LoadTemporal()
	if err != nil {
		return nil, fmt.Errorf("failed to load temporal config: %w", err)
	}

	config.Database, err = LoadDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	config.Kubernetes, err = LoadKubernetes()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
	}

	config.Worker, err = LoadWorker()
	if err != nil {
		return nil, fmt.Errorf("failed to load worker config: %w", err)
	}

	config.Timeouts, err = LoadTimeouts()
	if err != nil {
		return nil, fmt.Errorf("failed to load timeout config: %w", err)
	}

	config.Logging, err = LoadLogging()
	if err != nil {
		return nil, fmt.Errorf("failed to load logging config: %w", err)
	}

	logger.Info("Configuration loaded successfully")
	return config, nil
}

