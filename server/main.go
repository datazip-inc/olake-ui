package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/handlers"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/routes"
)

func main() {
	telemetry.InitTelemetry()
	constants.Init()
	logger.Init()

	// Initialize unified AppService
	logger.Info("Initializing application services...")
	appSvc, err := services.InitAppService()
	if err != nil {
		logger.Fatal("Failed to initialize services: %s", err)
		return
	}
	logger.Info("Application services initialized successfully")

	h := handlers.NewHandler(appSvc)
	routes.Init(h)
	if key := os.Getenv(constants.EncryptionKey); key == "" {
		logger.Warn("Encryption key is not set. This is not recommended for production environments.")
	}
	if web.BConfig.RunMode == "dev" || web.BConfig.RunMode == "staging" {
		orm.Debug = true
	}

	go web.Run()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	logger.Info("Shutting down server...")
}
