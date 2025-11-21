package constants

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/server/web"
	"github.com/spf13/viper"
)

var (
	defaultPort      = 8000
	version          = "Not Set"
	commitsha        = "Not Set"
	releasechannel   = "Not Set"
	defaultBaseHost  = "localhost"
	DefaultTimeZone  = "Asia/Kolkata"
	DefaultUsername  = "olake"
	DefaultPassword  = "password"
	EncryptionKey    = "OLAKE_SECRET_KEY"
	TableNameMap     = map[TableType]string{}
	DefaultConfigDir = "/tmp/olake-config"

	DefaultLogRetentionPeriod = 30
	DefaultCancelSyncWaitTime = 30 * time.Second

	// versions
	DefaultSpecVersion             = "v0.2.0"
	DefaultClearDestinationVersion = "v0.3.0"

	// logging
	EnvLogLevel          = "LOG_LEVEL"
	EnvLogFormat         = "LOG_FORMAT"
	OrderByUpdatedAtDesc = "-updated_at"
	// Frontend index path key
	FrontendIndexPath = "FRONTEND_INDEX_PATH"
	TemporalTaskQueue = "OLAKE_DOCKER_TASK_QUEUE"

	// conf keys
	ConfEncryptionKey         = "encryptionkey"
	ConfTemporalAddress       = "TEMPORAL_ADDRESS"
	ConfDeploymentMode        = "DEPLOYMENT_MODE"
	ConfRunMode               = "runmode"
	ConfContainerRegistryBase = "CONTAINER_REGISTRY_BASE"
	// database keys
	ConfPostgresDB            = "postgresdb"
	ConfOLakePostgresUser     = "OLAKE_POSTGRES_USER"
	ConfOLakePostgresPassword = "OLAKE_POSTGRES_PASSWORD"
	ConfOLakePostgresHost     = "OLAKE_POSTGRES_HOST"
	ConfOLakePostgresPort     = "OLAKE_POSTGRES_PORT"
	ConfOLakePostgresDBname   = "OLAKE_POSTGRES_DBNAME"
	ConfOLakePostgresSslmode  = "OLAKE_POSTGRES_SSLMODE"
)

// Supported database/source types
var SupportedSourceTypes = []string{
	"mysql",
	"postgres",
	"oracle",
	"mongodb",
	"kafka",
}

// Supported database/source types
var SupportedDestinationTypes = []string{
	"parquet",
	"iceberg",
}

var RequiredConfigVariable = []string{
	"OLAKE_POSTGRES_USER",
	"OLAKE_POSTGRES_PASSWORD",
	"OLAKE_POSTGRES_HOST",
	"OLAKE_POSTGRES_PORT",
	"OLAKE_POSTGRES_DBNAME",
	"OLAKE_POSTGRES_SSLMODE",
	"copyrequestbody",
	"logsdir"}

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault(EnvLogFormat, "console")
	viper.SetDefault(EnvLogLevel, "info")
	viper.SetDefault("PORT", defaultPort)
	viper.SetDefault("BUILD", version)
	viper.SetDefault("COMMITSHA", commitsha)
	viper.SetDefault("RELEASE_CHANNEL", releasechannel)
	viper.SetDefault("BASE_HOST", defaultBaseHost)
	viper.SetDefault("BASE_URL", fmt.Sprintf("%s:%v", viper.GetString("BASE_HOST"), viper.GetString("PORT")))
	viper.SetDefault(FrontendIndexPath, "/opt/frontend/dist/index.html")

	checkForRequiredVariables(RequiredConfigVariable)

	// init table names
	TableNameMap = map[TableType]string{
		UserTable:        "olake-$$-user",
		SourceTable:      "olake-$$-source",
		DestinationTable: "olake-$$-destination",
		JobTable:         "olake-$$-job",
		CatalogTable:     "olake-$$-catalog",
		SessionTable:     "session",
	}

	// replace $$ with the environment
	for k, v := range TableNameMap {
		TableNameMap[k] = strings.ReplaceAll(v, "$$", web.BConfig.RunMode)
	}
}

func checkForRequiredVariables(vars []string) {
	for _, v := range vars {
		value, err := config.String(v)
		if err != nil || value == "" {
			panic("Required config variable not found: ," + v)
		}
	}
}
