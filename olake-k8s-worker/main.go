package main

import (
	"os"
	"os/signal"
	"syscall"

	"olake-k8s-worker/config"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/worker"
)

func main() {
	// Initialize logger first
	logger.Init()

	logger.Info("OLake K8s Worker starting...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Apply environment-specific overrides
	cfg.ApplyEnvironmentOverrides()

	logger.Infof("Temporal Address: %s", cfg.Temporal.Address)
	logger.Infof("Task Queue: %s", cfg.Temporal.TaskQueue)
	logger.Infof("Namespace: %s", cfg.Kubernetes.Namespace)
	logger.Infof("Environment: %s", cfg.Database.RunMode)

	// Create K8s worker with configuration
	w, err := worker.NewK8sWorkerWithConfig(cfg)
	if err != nil {
		logger.Fatalf("Failed to create K8s worker: %v", err)
	}

	// Start worker
	if err := w.Start(); err != nil {
		logger.Fatalf("Failed to start worker: %v", err)
	}

	logger.Info("K8s Worker started successfully")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down K8s worker...")
	w.Stop()
	logger.Info("K8s worker stopped")
}
