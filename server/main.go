package main

import (
	"os"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/logger"
	"github.com/datazip/olake-frontend/server/routes"
)

func main() {
	// TODO: check if we have to create a new config file for docker compatibility
	if key := os.Getenv(constants.EncryptionKey); key == "" {
		logs.Warning("Encryption key is not set. This is not recommended for production environments.")
	}
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

	// init routers
	routes.Init()

	// setup environment mode
	if web.BConfig.RunMode == "dev" {
		orm.Debug = true
	}

	web.Run()
}
