package main

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/logger"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/routes"
)

func main() {
	// TODO: check if we have to create a new config file for docker compatibility
	go func() {
		if err := telemetry.InitTelemetry(); err != nil {
			logs.Error("Failed to initialize telemetry: %v", err)
		}
	}()

	defer telemetry.Flush()
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
		return
	}

	// init routers
	routes.Init()

	// setup environment mode
	if web.BConfig.RunMode == "dev" {
		orm.Debug = true
	}

	web.Run()
}
