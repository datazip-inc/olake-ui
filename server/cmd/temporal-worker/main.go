package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/datazip/olake-server/internal/temporal"
)

func main() {
	log.Println("Starting Olake Temporal worker...")

	// Default to localhost if no address provided
	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		temporalAddress = "localhost:7233"
	}

	// Create a new worker
	worker, err := temporal.NewWorker(temporalAddress)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	// Start the worker in a goroutine
	go func() {
		log.Printf("Starting worker with Temporal server at %s", temporalAddress)
		err := worker.Start()
		if err != nil {
			log.Fatalf("Failed to start worker: %v", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-signalChan
	log.Printf("Received signal %v, shutting down worker...", sig)

	// Stop the worker
	worker.Stop()
	log.Println("Worker stopped. Goodbye!")
}
