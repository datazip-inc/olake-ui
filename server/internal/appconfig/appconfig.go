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
}

var cfg = loadFromViper()

func Load() Config {
	return cfg
}

func loadFromViper() Config {
	v := viper.New()
	v.AutomaticEnv()

	v.SetDefault("APP_NAME", "olake-server")
	v.SetDefault("HTTP_PORT", "8000")
	v.SetDefault("RUN_MODE", "dev")
	v.SetDefault("COPY_REQUEST_BODY", true)
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

	return Config{
		AppName:               strings.TrimSpace(v.GetString("APP_NAME")),
		HTTPPort:              strings.TrimSpace(v.GetString("HTTP_PORT")),
		RunMode:               strings.TrimSpace(v.GetString("RUN_MODE")),
		CopyRequestBody:       v.GetBool("COPY_REQUEST_BODY"),
		MaxMemory:             v.GetInt64("MAX_MEMORY"),
		MaxUploadSize:         v.GetInt64("MAX_UPLOAD_SIZE"),
		PostgresDSN:           strings.TrimSpace(v.GetString("POSTGRES_DB")),
		EncryptionKey:         strings.TrimSpace(v.GetString("OLAKE_SECRET_KEY")),
		OlakePostgresUser:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_USER")),
		OlakePostgresPassword: strings.TrimSpace(v.GetString("OLAKE_POSTGRES_PASSWORD")),
		OlakePostgresHost:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_HOST")),
		OlakePostgresPort:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_PORT")),
		OlakePostgresDBName:   strings.TrimSpace(v.GetString("OLAKE_POSTGRES_DBNAME")),
		OlakePostgresSSLMode:  strings.TrimSpace(v.GetString("OLAKE_POSTGRES_SSLMODE")),
		LogsDir:               strings.TrimSpace(v.GetString("LOGS_DIR")),
		SessionOn:             v.GetBool("SESSION_ON"),
		TemporalAddress:       strings.TrimSpace(v.GetString("TEMPORAL_ADDRESS")),
		ContainerRegistryBase: strings.TrimSpace(v.GetString("CONTAINER_REGISTRY_BASE")),
	}
}
