package loader

import (
	"fmt"

	"olake-k8s-worker/config/types"
	"olake-k8s-worker/logger"
)

// ConfigLoader orchestrates the configuration loading process
type ConfigLoader struct {
	temporalProvider   TemporalProvider
	databaseProvider   DatabaseProvider
	kubernetesProvider KubernetesProvider
	workerProvider     WorkerProvider
	timeoutProvider    TimeoutProvider
	loggingProvider    LoggingProvider
}

// NewConfigLoader creates a new configuration loader with default providers
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		temporalProvider:   NewTemporalProvider(),
		databaseProvider:   NewDatabaseProvider(),
		kubernetesProvider: NewKubernetesProvider(),
		workerProvider:     NewWorkerProvider(),
		timeoutProvider:    NewTimeoutProvider(),
		loggingProvider:    NewLoggingProvider(),
	}
}

// LoadConfig loads the complete configuration
func (l *ConfigLoader) LoadConfig() (*types.Config, error) {
	config := &types.Config{}

	// Load each configuration section
	var err error
	
	config.Temporal, err = l.temporalProvider.LoadTemporal()
	if err != nil {
		return nil, fmt.Errorf("failed to load temporal config: %w", err)
	}

	config.Database, err = l.databaseProvider.LoadDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	config.Kubernetes, err = l.kubernetesProvider.LoadKubernetes()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
	}

	config.Worker, err = l.workerProvider.LoadWorker()
	if err != nil {
		return nil, fmt.Errorf("failed to load worker config: %w", err)
	}

	config.Timeouts, err = l.timeoutProvider.LoadTimeouts()
	if err != nil {
		return nil, fmt.Errorf("failed to load timeout config: %w", err)
	}

	config.Logging, err = l.loggingProvider.LoadLogging()
	if err != nil {
		return nil, fmt.Errorf("failed to load logging config: %w", err)
	}


	logger.Info("Configuration loaded successfully")
	return config, nil
}

