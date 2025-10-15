package main

import (
	"os"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
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
	// TODO: check if we have to create a new config file for docker compatibility
	if key := os.Getenv(constants.EncryptionKey); key == "" {
		logger.Warn("Encryption key is not set. This is not recommended for production environments.")
	}

	// start telemetry service
	telemetry.InitTelemetry()

	// check constants
	constants.Init()

	// init logger
	logger.Init()

	// init database
	err := database.Init()
	if err != nil {
		logger.Fatal("Failed to initialize database: %s", err)
		return
	}

	// Initialize all services once at startup
	// If any service fails to initialize, the application will exit immediately
	logger.Info("Initializing application services...")

	sourceService, err := services.NewSourceService()
	if err != nil {
		logger.Fatal("Failed to initialize SourceService: %s", err)
		return
	}
	logs.Info("SourceService initialized successfully")

	jobService, err := services.NewJobService()
	if err != nil {
		logger.Fatal("Failed to initialize JobService: %s", err)
		return
	}
	logger.Info("JobService initialized successfully")

	destinationService, err := services.NewDestinationService()
	if err != nil {
		logger.Fatal("Failed to initialize DestinationService: %s", err)
		return
	}
	logger.Info("DestinationService initialized successfully")

	userService := services.NewUserService()
	logger.Info("UserService initialized successfully")

	authService := services.NewAuthService()
	logger.Info("AuthService initialized successfully")

	// Initialize the app dependencies container
	handlers.InitApp(handlers.AppDeps{
		SourceService:      sourceService,
		JobService:         jobService,
		DestinationService: destinationService,
		UserService:        userService,
		AuthService:        authService,
	})
	logger.Info("Application dependencies initialized successfully")

	// init routers
	routes.Init()

	// setup environment mode
	if web.BConfig.RunMode == "dev" || web.BConfig.RunMode == "staging" {
		orm.Debug = true
	}

	web.Run()
}
