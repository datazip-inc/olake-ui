package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"github.com/datazip/olake-frontend/server/internal/logger"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/internal/temporal"
	"github.com/datazip/olake-frontend/server/utils"
)

func main() {
	// Initialize telemetry
	telemetry.InitTelemetry()
	// check constants
	constants.Init()

	// init logger
	logsdir, _ := config.String("logsdir")
	logger.InitLogger(logsdir)

	// init log cleaner
	utils.InitLogCleaner(docker.GetDefaultConfigDir(), utils.GetLogRetentionPeriod())

	// init database
	postgresDB, _ := config.String("postgresdb")
	err := database.Init(postgresDB)
	if err != nil {
		logs.Critical("Failed to initialize database: %s", err)
		return
	}

	logs.Info("Starting Olake Temporal worker...")

	// create temporal client
	tClient, err := temporal.NewClient()
	if err != nil {
		logs.Critical("Failed to create Temporal client: %v", err)
		return
	}
	defer tClient.Close()
	// create temporal worker
	worker, err := temporal.NewWorker(tClient)
	if err != nil {
		logs.Critical("Failed to create Temporal worker: %v", err)
		return
	}

	// Start the worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil {
			logs.Critical("Failed to start worker: %v", err)
			return
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
	telemetry.Close()
	logs.Info("Worker stopped. Goodbye!")
}
