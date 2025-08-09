package loader

import (
	"olake-ui/olake-workers/k8s/config/types"
	"olake-ui/olake-workers/k8s/utils/env"
	"olake-ui/olake-workers/k8s/utils/parser"
)

// LoadDatabase loads database configuration from environment variables
func LoadDatabase() (types.DatabaseConfig, error) {
	return types.DatabaseConfig{
		URL:             env.GetEnv("DATABASE_URL", ""),
		Host:            env.GetEnv("DB_HOST", "postgres.olake.svc.cluster.local"),
		Port:            env.GetEnv("DB_PORT", "5432"),
		User:            env.GetEnv("DB_USER", "postgres"),
		Password:        env.GetEnv("DB_PASSWORD", "password"),
		Database:        env.GetEnv("DB_NAME", "olake"),
		SSLMode:         env.GetEnv("DB_SSLMODE", "disable"),
		RunMode:         env.GetEnv("RUN_MODE", "dev"),
		MaxOpenConns:    env.GetEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    env.GetEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: parser.ParseDuration("DB_CONN_MAX_LIFETIME", "5m"),
	}, nil
}
