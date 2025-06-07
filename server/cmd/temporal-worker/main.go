package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/logger"
	"github.com/datazip/olake-server/internal/temporal"
)

func main() {
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
	}

	logs.Info("Starting Olake Temporal worker...")

	// Create a new worker
	worker, err := temporal.NewWorker()
	if err != nil {
		logs.Critical("Failed to create worker: %v", err)
	}

	// Start the worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil {
			logs.Critical("Failed to start worker: %v", err)
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
