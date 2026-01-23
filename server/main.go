// @title Olake UI API
// @version 0.2.9
// @description Olake is fastest open-source tool for replicating Databases to Apache Iceberg or Data Lakehouse.
// @contact.name OLake
// @contact.email hello@olake.io
// @contact.url https://olake.io/contact/
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @tag.name Authentication
// @tag.description Authentication endpoints
// @tag.name Jobs
// @tag.description Job management and execution endpoints
// @tag.name Sources
// @tag.description Source configuration endpoints
// @tag.name Destinations
// @tag.description Destination configuration endpoints
// @tag.name Project Settings
// @tag.description Project configuration endpoints
// @tag.name Platform
// @tag.description Platform-level operations
// @tag.name Users
// @tag.description User management endpoints
// @tag.name Internal
// @tag.description Internal worker callbacks (not for external use)

package main

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/routes"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
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
	if key, _ := web.AppConfig.String(constants.ConfEncryptionKey); key == "" {
		logger.Warn("Encryption key is not set. This is not recommended for production environments.")
	}

	if web.BConfig.RunMode == "dev" || web.BConfig.RunMode == "staging" {
		orm.Debug = true
	}
	web.Run()
	// TODO: handle gracefull shutdown
}
