package main

import (
	appConfig "olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/worker"
)

func main() {
	// Load full configuration (including logging config)
	cfg, err := appConfig.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize logger with loaded configuration
	logger.Init(&cfg.Logging)

	logger.Infof("Temporal Address: %s", cfg.Temporal.Address)
	logger.Infof("Task Queue: %s", cfg.Temporal.TaskQueue)
	logger.Infof("Namespace: %s", cfg.Kubernetes.Namespace)
	logger.Infof("Environment: %s", cfg.Database.RunMode)

	// Create K8s worker with configuration
	w, err := worker.NewK8sWorker(cfg)
	if err != nil {
		logger.Fatalf("Failed to create K8s worker: %v", err)
	}

	// Start worker - this will block until shutdown signal is received
	if err := w.Start(); err != nil {
		logger.Fatalf("Failed to start worker: %v", err)
	}
}
