package config

import (
	"github.com/spf13/viper"
)

var globalViper *viper.Viper

// InitConfig initializes the global Viper instance
func InitConfig() {
	globalViper = viper.New()
	globalViper.AutomaticEnv()
}

// LoadConfig loads configuration using Viper from environment variables
func LoadConfig() (*Config, error) {
	if globalViper == nil {
		InitConfig()
	}

	// Bind environment variables to structured config
	bindEnvironmentVariables(globalViper)

	// Unmarshal into struct
	var config Config
	if err := globalViper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// bindEnvironmentVariables binds environment variables to config paths
func bindEnvironmentVariables(v *viper.Viper) {
	// Database bindings
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.database", "DB_NAME")
	v.BindEnv("database.ssl_mode", "DB_SSLMODE")
	v.BindEnv("database.run_mode", "RUN_MODE")
	v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	// Temporal bindings
	v.BindEnv("temporal.address", "TEMPORAL_ADDRESS")
	v.BindEnv("temporal.task_queue", "TEMPORAL_TASK_QUEUE")
	v.BindEnv("temporal.timeout", "TEMPORAL_TIMEOUT")

	// Kubernetes bindings
	v.BindEnv("kubernetes.namespace", "WORKER_NAMESPACE")
	v.BindEnv("kubernetes.image_registry", "IMAGE_REGISTRY")
	v.BindEnv("kubernetes.image_pull_policy", "IMAGE_PULL_POLICY")
	v.BindEnv("kubernetes.service_account", "SERVICE_ACCOUNT")
	v.BindEnv("kubernetes.storage_pvc_name", "OLAKE_STORAGE_PVC_NAME")
	v.BindEnv("kubernetes.job_service_account", "JOB_SERVICE_ACCOUNT_NAME")
	v.BindEnv("kubernetes.secret_key", "OLAKE_SECRET_KEY")
	v.BindEnv("kubernetes.labels.version", "WORKER_VERSION")

	// Worker bindings
	v.BindEnv("worker.max_concurrent_activities", "MAX_CONCURRENT_ACTIVITIES")
	v.BindEnv("worker.max_concurrent_workflows", "MAX_CONCURRENT_WORKFLOWS")
	v.BindEnv("worker.heartbeat_interval", "HEARTBEAT_INTERVAL")
	v.BindEnv("worker.worker_identity", "POD_NAME")

	// Timeout bindings
	v.BindEnv("timeouts.activity.discover", "TIMEOUT_ACTIVITY_DISCOVER")
	v.BindEnv("timeouts.activity.test", "TIMEOUT_ACTIVITY_TEST")
	v.BindEnv("timeouts.activity.sync", "TIMEOUT_ACTIVITY_SYNC")

	// Logging bindings
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")
}
