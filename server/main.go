package main

import (
	"os"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
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

	err := database.Init()
	if err != nil {
		logger.Fatal("Failed to initialize database: %s", err)
		return
	}

	// Initialize all services at once
	logger.Info("Initializing application services...")
	svcs, err := services.InitServices()
	if err != nil {
		logger.Fatal("Failed to initialize services: %v", err)
		return
	}
	logger.Info("Application services initialized successfully")

	// Initialize handlers with services
	handlers.InitHandlers(svcs)

	routes.Init()
	if key := os.Getenv(constants.EncryptionKey); key == "" {
		logger.Warn("Encryption key is not set. This is not recommended for production environments.")
	}
	if web.BConfig.RunMode == "dev" || web.BConfig.RunMode == "staging" {
		orm.Debug = true
	}

	web.Run()
}
