package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/rest"

	"olake-k8s-worker/shared"
	"olake-k8s-worker/worker"
)

func main() {
	fmt.Println("OLake K8s Worker - Basic Setup Test")

	// Try to create a basic k8s config (this will fail outside cluster, but tests imports)
	_, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Not running in cluster (expected): %v", err)
	}

	fmt.Println("K8s imports working!")

	log.Println("Starting OLake K8s Worker...")

	// Log configuration
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddr == "" {
		temporalAddr = shared.DefaultTemporalAddress
	}

	log.Printf("Temporal Address: %s", temporalAddr)
	log.Printf("Task Queue: %s", shared.TaskQueue)

	// Create K8s worker
	w, err := worker.NewK8sWorker()
	if err != nil {
		log.Fatalf("Failed to create K8s worker: %v", err)
	}

	// Start worker
	if err := w.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	log.Println("K8s Worker started successfully")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down K8s worker...")
	w.Stop()
	log.Println("K8s worker stopped")
}
