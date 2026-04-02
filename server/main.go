// @title Olake UI API
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
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/datazip-inc/olake-ui/server/internal/httpserver"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"

	"github.com/datazip-inc/olake-ui/server/docs"
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

	// Set Swagger Info version to match the application's runtime version.
	if constants.AppVersion != "" {
		docs.SwaggerInfo.Version = constants.AppVersion
	}

	cfg := appconfig.Load()
	if cfg.EncryptionKey == "" {
		logger.Warn("Encryption key is not set. This is not recommended for production environments.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api, err := handlers.NewHandler(appSvc, &cfg, db)
	if err != nil {
		logger.Fatalf("Failed to initialize handler: %s", err)
		return
	}
	server := httpserver.New(&cfg, api)
	logger.Infof("Starting HTTP server on port %s", cfg.HTTPPort)
	if err := server.Run(ctx); err != nil {
		logger.Fatalf("HTTP server exited with error: %s", err)
	}
}
