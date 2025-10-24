package main

import (
	"os"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/etl-service"
	"github.com/datazip/olake-ui/server/internal/handlers"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/routes"
)

func main() {
	constants.Init()
	logger.Init()
	db, err := database.Init()
	if err != nil {
		logger.Fatalf("Failed to initialize database: %s", err)
		return
	}

	// Initialize unified AppService
	appSvc, err := services.InitAppService(db)
	if err != nil {
		logger.Fatalf("Failed to initialize services: %s", err)
		return
	}
	logger.Info("Application services initialized successfully")
	telemetry.InitTelemetry(db)

	routes.Init(handlers.NewHandler(appSvc))
	if key := os.Getenv(constants.EncryptionKey); key == "" {
		logger.Warn("Encryption key is not set. This is not recommended for production environments.")
	}

	if web.BConfig.RunMode == "dev" || web.BConfig.RunMode == "staging" {
		orm.Debug = true
	}
	web.Run()
	// TODO: handle gracefull shutdown
}
