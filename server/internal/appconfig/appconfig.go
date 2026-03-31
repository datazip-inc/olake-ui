package appconfig

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppName               string
	HTTPPort              string
	RunMode               string
	CopyRequestBody       bool
	MaxMemory             int64
	MaxUploadSize         int64
	PostgresDSN           string
	EncryptionKey         string
	OlakePostgresUser     string
	OlakePostgresPassword string
	OlakePostgresHost     string
	OlakePostgresPort     string
	OlakePostgresDBName   string
	OlakePostgresSSLMode  string
	LogsDir               string
	SessionOn             bool
	TemporalAddress       string
	ContainerRegistryBase string
	EnableOptimization    bool
	OptimizationGroup     string
	OptimizationBaseURL   string
	OptimizationUsername  string
	OptimizationPassword  string
}

var cfg = loadConfig()

func Load() Config {
	return cfg
}

func loadConfig() Config {
	v := viper.New()

	// Note: config priority: env variables -> file (app.yaml) -> default values

	// set default values
	v.SetDefault("APP_NAME", "olake-server")
	v.SetDefault("HTTP_PORT", "8000")
	v.SetDefault("RUN_MODE", "dev")
	v.SetDefault("MAX_MEMORY", int64(67108864))
	v.SetDefault("MAX_UPLOAD_SIZE", int64(67108864))
	v.SetDefault("OLAKE_POSTGRES_USER", "temporal")
	v.SetDefault("OLAKE_POSTGRES_PASSWORD", "temporal")
	v.SetDefault("OLAKE_POSTGRES_HOST", "postgresql")
	v.SetDefault("OLAKE_POSTGRES_PORT", "5432")
	v.SetDefault("OLAKE_POSTGRES_DBNAME", "postgres")
	v.SetDefault("OLAKE_POSTGRES_SSLMODE", "disable")
	v.SetDefault("LOGS_DIR", "./logger/logs")
	v.SetDefault("SESSION_ON", true)
	v.SetDefault("TEMPORAL_ADDRESS", "temporal:7233")
	v.SetDefault("CONTAINER_REGISTRY_BASE", "registry-1.docker.io")
	v.SetDefault("OPTIMIZATION_GROUP", "spark-container")

	v.SetConfigFile("./conf/app.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	// load environment variables
	v.AutomaticEnv()

	return Config{
		AppName:               strings.TrimSpace(v.GetString("APP_NAME")),
		HTTPPort:              strings.TrimSpace(v.GetString("HTTP_PORT")),
		RunMode:               strings.TrimSpace(v.GetString("RUN_MODE")),
		ContainerRegistryBase: strings.TrimSpace(v.GetString("CONTAINER_REGISTRY_BASE")),
		LogsDir:               strings.TrimSpace(v.GetString("LOGS_DIR")),
		TemporalAddress:       strings.TrimSpace(v.GetString("TEMPORAL_ADDRESS")),
		CopyRequestBody:       v.GetBool("COPY_REQUEST_BODY"),
		MaxMemory:             v.GetInt64("MAX_MEMORY"),
		MaxUploadSize:         v.GetInt64("MAX_UPLOAD_SIZE"),
		SessionOn:             v.GetBool("SESSION_ON"),

		PostgresDSN:           strings.TrimSpace(v.GetString("POSTGRES_DB")),
		EncryptionKey:         strings.TrimSpace(v.GetString("OLAKE_SECRET_KEY")),
		OlakePostgresUser:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_USER")),
		OlakePostgresPassword: strings.TrimSpace(v.GetString("OLAKE_POSTGRES_PASSWORD")),
		OlakePostgresHost:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_HOST")),
		OlakePostgresPort:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_PORT")),
		OlakePostgresDBName:   strings.TrimSpace(v.GetString("OLAKE_POSTGRES_DBNAME")),
		OlakePostgresSSLMode:  strings.TrimSpace(v.GetString("OLAKE_POSTGRES_SSLMODE")),

		EnableOptimization:   v.GetBool("ENABLE_OPTIMIZATION"),
		OptimizationGroup:    strings.TrimSpace(v.GetString("OPTIMIZATION_GROUP")),
		OptimizationBaseURL:  strings.TrimSpace(v.GetString("OPTIMIZATION_BASE_URL")),
		OptimizationUsername: strings.TrimSpace(v.GetString("USERNAME")),
		OptimizationPassword: strings.TrimSpace(v.GetString("PASSWORD")),
	}
}
