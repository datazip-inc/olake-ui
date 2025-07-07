package main

import (
	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/worker"
)

func main() {
	// Initialize bootstrap logger for early startup messages
	logger.InitDefault()
	
	logger.Info("OLake K8s Worker starting...")
	logger.Info("Loading configuration...")

	// Load full configuration (including logging config)
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Re-initialize logger with loaded configuration
	logger.Init(cfg.Logging)
	logger.Info("Logger reconfigured with loaded settings")


	logger.Infof("Temporal Address: %s", cfg.Temporal.Address)
	logger.Infof("Task Queue: %s", cfg.Temporal.TaskQueue)
	logger.Infof("Namespace: %s", cfg.Kubernetes.Namespace)
	logger.Infof("Environment: %s", cfg.Database.RunMode)

	// Create K8s worker with configuration
	w, err := worker.NewK8sWorkerWithConfig(cfg)
	if err != nil {
		logger.Fatalf("Failed to create K8s worker: %v", err)
	}

	// Start worker - this will block until shutdown signal is received
	if err := w.Start(); err != nil {
		logger.Fatalf("Failed to start worker: %v", err)
	}
}
