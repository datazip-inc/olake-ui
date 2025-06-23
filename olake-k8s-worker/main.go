package main

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/rest"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/worker"
)

func main() {
	// Initialize logger first
	logger.Init()

	logger.Info("OLake K8s Worker - Basic Setup Test")

	// Try to create a basic k8s config (this will fail outside cluster, but tests imports)
	_, err := rest.InClusterConfig()
	if err != nil {
		logger.Warnf("Not running in cluster (expected): %v", err)
	}

	logger.Info("K8s imports working!")

	logger.Info("Starting OLake K8s Worker...")

	// Log configuration
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddr == "" {
		temporalAddr = shared.DefaultTemporalAddress
	}

	logger.Infof("Temporal Address: %s", temporalAddr)
	logger.Infof("Task Queue: %s", shared.TaskQueue)

	// Create K8s worker
	w, err := worker.NewK8sWorker()
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
