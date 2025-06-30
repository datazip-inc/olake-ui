package loader

import (
	"fmt"
	"os"
	"time"

	"olake-k8s-worker/config/types"
	"olake-k8s-worker/logger"
	"olake-k8s-worker/utils/env"
)

// Provider interfaces for different configuration sections

type TemporalProvider interface {
	LoadTemporal() (types.TemporalConfig, error)
}

type DatabaseProvider interface {
	LoadDatabase() (types.DatabaseConfig, error)
}

type KubernetesProvider interface {
	LoadKubernetes() (types.KubernetesConfig, error)
}

type WorkerProvider interface {
	LoadWorker() (types.WorkerConfig, error)
}

type TimeoutProvider interface {
	LoadTimeouts() (types.TimeoutConfig, error)
}

type LoggingProvider interface {
	LoadLogging() (types.LoggingConfig, error)
}

// Default provider implementations

type defaultTemporalProvider struct{}
type defaultDatabaseProvider struct{}
type defaultKubernetesProvider struct{}
type defaultWorkerProvider struct{}
type defaultTimeoutProvider struct{}
type defaultLoggingProvider struct{}

func NewTemporalProvider() TemporalProvider {
	return &defaultTemporalProvider{}
}

func NewDatabaseProvider() DatabaseProvider {
	return &defaultDatabaseProvider{}
}

func NewKubernetesProvider() KubernetesProvider {
	return &defaultKubernetesProvider{}
}

func NewWorkerProvider() WorkerProvider {
	return &defaultWorkerProvider{}
}

func NewTimeoutProvider() TimeoutProvider {
	return &defaultTimeoutProvider{}
}

func NewLoggingProvider() LoggingProvider {
	return &defaultLoggingProvider{}
}

// Provider implementations

func (p *defaultTemporalProvider) LoadTemporal() (types.TemporalConfig, error) {
	return types.TemporalConfig{
		Address:   env.GetEnv("TEMPORAL_ADDRESS", "temporal.olake.svc.cluster.local:7233"),
		TaskQueue: "OLAKE_K8S_TASK_QUEUE", // Hardcoded as per requirement
		Timeout:   ParseDuration("TEMPORAL_TIMEOUT", "30s"),
	}, nil
}

func (p *defaultDatabaseProvider) LoadDatabase() (types.DatabaseConfig, error) {
	return types.DatabaseConfig{
		URL:             env.GetEnv("DATABASE_URL", ""),
		Host:            env.GetEnv("DB_HOST", "postgres.olake.svc.cluster.local"),
		Port:            env.GetEnv("DB_PORT", "5432"),
		User:            env.GetEnv("DB_USER", "postgres"),
		Password:        env.GetEnv("DB_PASSWORD", "password"),
		Database:        env.GetEnv("DB_NAME", "olake"),
		SSLMode:         env.GetEnv("DB_SSLMODE", "disable"),
		RunMode:         env.GetEnv("RUN_MODE", "production"),
		MaxOpenConns:    env.GetEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    env.GetEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: ParseDuration("DB_CONN_MAX_LIFETIME", "5m"),
	}, nil
}

func (p *defaultKubernetesProvider) LoadKubernetes() (types.KubernetesConfig, error) {
	return types.KubernetesConfig{
		Namespace:       env.GetEnv("WORKER_NAMESPACE", "olake"),
		ImageRegistry:   env.GetEnv("IMAGE_REGISTRY", "olakego"),
		ImagePullPolicy: env.GetEnv("IMAGE_PULL_POLICY", "IfNotPresent"),
		ServiceAccount:  env.GetEnv("SERVICE_ACCOUNT", "olake-worker"),
		PVCName:         env.GetEnv("OLAKE_STORAGE_PVC_NAME", "olake-jobs-pvc"),
		JobTTL:          getOptionalTTL("JOB_TTL_SECONDS", 0),
		JobTimeout:      ParseDuration("JOB_TIMEOUT", "15m"),
		CleanupPolicy:   env.GetEnv("CLEANUP_POLICY", "auto"),
		Labels: map[string]string{
			"app":        "olake-sync",
			"managed-by": "olake-k8s-worker",
			"version":    env.GetEnv("WORKER_VERSION", "latest"),
		},
	}, nil
}

func (p *defaultWorkerProvider) LoadWorker() (types.WorkerConfig, error) {
	return types.WorkerConfig{
		MaxConcurrentActivities: env.GetEnvInt("MAX_CONCURRENT_ACTIVITIES", 10),
		MaxConcurrentWorkflows:  env.GetEnvInt("MAX_CONCURRENT_WORKFLOWS", 5),
		HeartbeatInterval:       ParseDuration("HEARTBEAT_INTERVAL", "5s"),
		WorkerIdentity:          generateWorkerIdentity(),
	}, nil
}

func (p *defaultTimeoutProvider) LoadTimeouts() (types.TimeoutConfig, error) {
	return types.TimeoutConfig{
		WorkflowExecution: types.WorkflowTimeouts{
			Discover: ParseDuration("WORKFLOW_TIMEOUT_DISCOVER", "2h"), // 2 hours for discovery workflows
			Test:     ParseDuration("WORKFLOW_TIMEOUT_TEST", "2h"),     // 2 hours for test workflows
			Sync:     ParseDuration("WORKFLOW_TIMEOUT_SYNC", "720h"),   // 30 days for sync workflows
		},
		Activity: types.ActivityTimeouts{
			Discover: ParseDuration("ACTIVITY_TIMEOUT_DISCOVER", "30m"), // 30 minutes for discovery activities
			Test:     ParseDuration("ACTIVITY_TIMEOUT_TEST", "30m"),     // 30 minutes for test activities
			Sync:     ParseDuration("ACTIVITY_TIMEOUT_SYNC", "700h"),    // 29 days for sync activities (effectively infinite)
		},
	}, nil
}

func (p *defaultLoggingProvider) LoadLogging() (types.LoggingConfig, error) {
	return types.LoggingConfig{
		Level:      env.GetEnv("LOG_LEVEL", "info"),
		Format:     env.GetEnv("LOG_FORMAT", "console"),
		Structured: env.GetEnvBool("LOG_STRUCTURED", false),
	}, nil
}

// Helper functions

func generateWorkerIdentity() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("olake-k8s-worker-%s", hostname)
}

func getOptionalTTL(envKey string, defaultValue int) *int32 {
	value := env.GetEnvInt(envKey, defaultValue)
	if value <= 0 {
		return nil
	}
	ttl := int32(value)
	return &ttl
}

// ParseDuration parses a duration string with error handling and fallback
func ParseDuration(envKey, defaultValue string) time.Duration {
	value := env.GetEnv(envKey, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		logger.Warnf("Failed to parse duration for %s, using default: %s", envKey, defaultValue)
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}
