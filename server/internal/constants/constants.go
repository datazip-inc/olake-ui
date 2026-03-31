package constants

import (
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
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

	DefaultLogRetentionPeriod   = 30
	DefaultCancelSyncWaitTime   = 30 * time.Second
	DefaultListWorkflowPageSize = 500

	// versions
	DefaultSpecVersion               = "v0.2.0"
	DefaultClearDestinationVersion   = "v0.3.0"
	DefaultMaxDiscoverThreadsVersion = "v0.3.18"

	// logging
	EnvLogLevel          = "LOG_LEVEL"
	EnvLogFormat         = "LOG_FORMAT"
	OrderByUpdatedAtDesc = "-updated_at"
	// Frontend index path key
	FrontendIndexPath = "FRONTEND_INDEX_PATH"
	TemporalTaskQueue = "OLAKE_DOCKER_TASK_QUEUE"

	// command flags
	MaxDiscoverThreadsFlag    = "--max-discover-threads"
	DefaultMaxDiscoverThreads = 50

	// conf keys
	ConfEncryptionKey         = "encryptionkey"
	ConfTemporalAddress       = "TEMPORAL_ADDRESS"
	ConfDeploymentMode        = "DEPLOYMENT_MODE"
	ConfRunMode               = "runmode"
	ConfContainerRegistryBase = "CONTAINER_REGISTRY_BASE"
	ConfOptimizationBaseURL   = "OPTIMIZATION_BASE_URL"
	ConfOptimizationUsername  = "USERNAME"
	ConfOptimizationPassword  = "PASSWORD"
	ConfOptimizationGroup     = "OPTIMIZATION_GROUP"

	// database keys
	ConfPostgresDB            = "postgresdb"
	ConfOLakePostgresUser     = "OLAKE_POSTGRES_USER"
	ConfOLakePostgresPassword = "OLAKE_POSTGRES_PASSWORD"
	ConfOLakePostgresHost     = "OLAKE_POSTGRES_HOST"
	ConfOLakePostgresPort     = "OLAKE_POSTGRES_PORT"
	ConfOLakePostgresDBname   = "OLAKE_POSTGRES_DBNAME"
	ConfOLakePostgresSslmode  = "OLAKE_POSTGRES_SSLMODE"

	// Optimization API paths
	OptPathCatalogs                 = "/api/ams/v1/catalogs"
	OptPathCatalogDetail            = "/api/ams/v1/catalogs/%s"
	OptPathCatalogTables            = "/api/ams/v1/catalogs/%s/databases/%s/tables"
	OptPathTableDetails             = "/api/ams/v1/tables/catalogs/%s/dbs/%s/tables/%s/details"
	OptPathTableOptimizingProcesses = "/api/ams/v1/tables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes"
	OptPathTerminalExecute          = "/api/ams/v1/terminal/catalogs/%s/execute"
	OptPathTerminalLogs             = "/api/ams/v1/terminal/%s/logs"

	OptMaxTimeout = 30 * time.Second
	PollInterval  = 1500 * time.Millisecond

	OptMinorCron          = "self-optimizing.minor.trigger.cron"
	OptMajorCron          = "self-optimizing.major.trigger.cron"
	OptFullCron           = "self-optimizing.full.trigger.cron"
	OptTargetFileSize     = "write.target-file-size-bytes"
	OptEnableOptimization = "self-optimizing.enabled"

	OptSQLCommand = "ALTER TABLE %s.%s SET TBLPROPERTIES (%s)"

	// app env
	EnvAppEnvironment    = "APP_ENV"
	EnvCustomDriverImage = "CUSTOM_DRIVER_VERSION"

	// App environment supported values: production/development
	AppEnvProduction  = "production"
	AppEnvDevelopment = "development"

	// logs config
	// LogReadChunkSize is the number of bytes read per chunk when scanning log files.
	LogReadChunkSize = 64 * 1024 // 64KB

	// DefaultLogsLimit is the number of log entries returned if no limit is provided.
	DefaultLogsLimit = 1000

	// DefaultLogsCursor indicates tailing from the end of the file (cursor < 0).
	DefaultLogsCursor int64 = -1

	// DefaultLogsDirection is the fallback pagination direction ("older" or "newer").
	DefaultLogsDirection = "older"

	// ExecutorEnvironment indicates the runtime environment. Defaults to "docker"
	// and is updated to "kubernetes" at startup if KUBERNETES_SERVICE_HOST is set.
	ExecutorEnvironment = "docker"

	// TableFormatList defines supported table formats for catalogs
	TableFormatList = []string{"ICEBERG"}

	// hard-coding to S3 now, as the other options are "hadoop" & "OSS" for optimization
	// GCS & ADLS are supported, given the catalog manages the sdk (eg, Lakekeeper with GCS flavour)
	DefaultStroageType = "S3"
)

// Supported database/source types
var SupportedSourceTypes = []string{
	"mysql",
	"postgres",
	"oracle",
	"mongodb",
	"kafka",
	"s3",
	"db2",
	"mssql",
}

// Supported database/source types
var SupportedDestinationTypes = []string{
	"parquet",
	"iceberg",
}

var AppVersion string

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault(EnvLogFormat, "console")
	viper.SetDefault(EnvLogLevel, "info")
	viper.SetDefault(EnvAppEnvironment, AppEnvProduction)
	viper.SetDefault("PORT", defaultPort)
	viper.SetDefault("BUILD", version)
	viper.SetDefault("COMMITSHA", commitsha)
	viper.SetDefault("RELEASE_CHANNEL", releasechannel)
	viper.SetDefault("BASE_HOST", defaultBaseHost)
	viper.SetDefault("BASE_URL", fmt.Sprintf("%s:%v", viper.GetString("BASE_HOST"), viper.GetString("PORT")))
	viper.SetDefault(FrontendIndexPath, "/opt/frontend/dist/index.html")

	checkForRequiredVariables()

	AppVersion = viper.GetString("APP_VERSION")

	if viper.GetString("KUBERNETES_SERVICE_HOST") != "" {
		ExecutorEnvironment = "kubernetes"
	}

	// init table names
	TableNameMap = map[TableType]string{
		UserTable:            "olake-$$-user",
		SourceTable:          "olake-$$-source",
		DestinationTable:     "olake-$$-destination",
		JobTable:             "olake-$$-job",
		CatalogTable:         "olake-$$-catalog",
		SessionTable:         "session",
		ProjectSettingsTable: "olake-$$-project-settings",
	}

	// replace $$ with the environment
	for k, v := range TableNameMap {
		TableNameMap[k] = strings.ReplaceAll(v, "$$", appconfig.Load().RunMode)
	}
}

func checkForRequiredVariables() {
	cfg := appconfig.Load()

	// If a full DSN is provided, we don't require individual DB parts.
	if strings.TrimSpace(cfg.PostgresDSN) != "" {
		return
	}

	requiredValues := map[string]string{
		"OLAKE_POSTGRES_USER":     cfg.OlakePostgresUser,
		"OLAKE_POSTGRES_PASSWORD": cfg.OlakePostgresPassword,
		"OLAKE_POSTGRES_HOST":     cfg.OlakePostgresHost,
		"OLAKE_POSTGRES_PORT":     cfg.OlakePostgresPort,
		"OLAKE_POSTGRES_DBNAME":   cfg.OlakePostgresDBName,
		"OLAKE_POSTGRES_SSLMODE":  cfg.OlakePostgresSSLMode,
	}

	for name, value := range requiredValues {
		if strings.TrimSpace(value) == "" {
			panic("Required config variable not found: " + name)
		}
	}
}
