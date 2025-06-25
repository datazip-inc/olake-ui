package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/logger"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/internal/temporal"
)

func main() {
	// Initialize telemetry
	go func() {
		if err := telemetry.InitTelemetry(); err != nil {
			logs.Error("Failed to initialize telemetry: %v", err)
		}
	}()
	defer time.Sleep(5 * time.Second)

	// check constants
	constants.Init()

	// init logger
	logsdir, _ := config.String("logsdir")
	logger.InitLogger(logsdir)
	// init database
	postgresDB, _ := config.String("postgresdb")
	err := database.Init(postgresDB)
	if err != nil {
		logs.Critical("Failed to initialize database: %s", err)
		os.Exit(1)
	}

	logs.Info("Starting Olake Temporal worker...")

	// Create a new worker
	worker, err := temporal.NewWorker()
	if err != nil {
		logs.Critical("Failed to create worker: %v", err)
		os.Exit(1)
	}

	// Start the worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil {
			logs.Critical("Failed to start worker: %v", err)
			os.Exit(1)
		}
	}()

	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-signalChan
	logs.Info("Received signal %v, shutting down worker...", sig)

	// Stop the worker
	worker.Stop()
	logs.Info("Worker stopped. Goodbye!")
}
