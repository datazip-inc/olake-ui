package appconfig

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName                        string
	HTTPPort                       string
	HTTPReadHeaderTimeout          time.Duration
	HTTPReadTimeout                time.Duration
	HTTPWriteTimeout               time.Duration
	HTTPIdleTimeout                time.Duration
	RunMode                        string
	MaxMemory                      int64
	MaxUploadSize                  int64
	PostgresDSN                    string
	EncryptionKey                  string
	OlakePostgresUser              string
	OlakePostgresPassword          string
	OlakePostgresHost              string
	OlakePostgresPort              string
	OlakePostgresDBName            string
	OlakePostgresSSLMode           string
	LogsDir                        string
	SessionOn                      bool
	TemporalAddress                string
	TemporalNamespace              string
	TemporalEnableTLS              bool
	TemporalAPIKey                 string
	TemporalExternal               bool
	TemporalTaskQueue              string
	ContainerRegistryBase          string
	ContainerRegistryUsername      string
	ContainerRegistryPassword      string
	ContainerRegistryInsecure      bool
	ContainerRegistryTLSSkipVerify bool
	ContainerRegistryCACert        string
	MetricsEnabled                 bool
	EnableOptimization             bool
	OptimizationGroup              string
	OptimizationBaseURL            string
	OptimizationUsername           string
	OptimizationPassword           string
	OptimizationRequestTimeout     time.Duration
	GitOpsMode                     string // "kubernetes" or "file"; empty disables GitOps
	PodNamespace                   string // namespace Kubernetes mode watches
	GitOpsFileDir                  string // file based configuration for olake; defaults to /etc/olake/gitops
	GitOpsStateFile                string // file mode status; defaults to OLAKE_GITOPS_FILE_DIR/.olake-gitops-state.json
	GitOpsSecretsRefreshInterval   time.Duration
	GitOpsSecretsAWSRegion         string
	GitOpsSecretsAWSAccessKeyID    string
	GitOpsSecretsAWSSecretKey      string
	GitOpsSecretsGCPProject        string
	GitOpsSecretsAzureTenantID     string
	GitOpsSecretsAzureClientID     string
	GitOpsSecretsAzureClientSecret string
	OlakeStorageMode               string
	OlakeS3Bucket                  string
	OlakeS3Region                  string
	OlakeS3Prefix                  string
	OlakeS3AccessKeyID             string
	OlakeS3SecretAccessKey         string
	OlakeS3Endpoint                string
}

// load reads config/app.yaml and the environment once, on first use, so importing this package
// has no side effects.
var load = sync.OnceValue(loadConfig)

func Load() Config {
	return load()
}

func loadConfig() Config {
	v := viper.New()

	// Auto-detect RUN_MODE default based on execution environment.
	// KUBERNETES_SERVICE_HOST is automatically injected into every pod by Kubernetes.
	// Priority: RUN_MODE env var > app.yaml > this default.
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		v.SetDefault("RUN_MODE", "staging")
	} else {
		v.SetDefault("RUN_MODE", "dev")
	}

	// Prometheus /metrics is on by default; METRICS_ENABLED=false turns the route off.
	v.SetDefault("METRICS_ENABLED", true)

	// GitOps file mode reads manifests from here; mount them at this path.
	v.SetDefault("OLAKE_GITOPS_FILE_DIR", "/etc/olake/gitops")

	// Resources whose config has secret references are re-checked this often for rotated secrets.
	v.SetDefault("GITOPS_SECRETS_REFRESH_INTERVAL", "1h")

	// Note: config priority: env variables -> file (app.yaml)
	v.SetConfigFile("./config/app.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	// load environment variables
	v.AutomaticEnv()

	return Config{
		AppName:                        strings.TrimSpace(v.GetString("APP_NAME")),
		HTTPPort:                       strings.TrimSpace(v.GetString("HTTP_PORT")),
		HTTPReadHeaderTimeout:          v.GetDuration("HTTP_READ_HEADER_TIMEOUT"),
		HTTPReadTimeout:                v.GetDuration("HTTP_READ_TIMEOUT"),
		HTTPWriteTimeout:               v.GetDuration("HTTP_WRITE_TIMEOUT"),
		HTTPIdleTimeout:                v.GetDuration("HTTP_IDLE_TIMEOUT"),
		RunMode:                        strings.TrimSpace(v.GetString("RUN_MODE")),
		ContainerRegistryBase:          strings.TrimSpace(v.GetString("CONTAINER_REGISTRY_BASE")),
		ContainerRegistryUsername:      strings.TrimSpace(v.GetString("CONTAINER_REGISTRY_USERNAME")),
		ContainerRegistryPassword:      strings.TrimSpace(v.GetString("CONTAINER_REGISTRY_PASSWORD")),
		ContainerRegistryInsecure:      v.GetBool("CONTAINER_REGISTRY_INSECURE"),
		ContainerRegistryTLSSkipVerify: v.GetBool("CONTAINER_REGISTRY_TLS_SKIP_VERIFY"),
		ContainerRegistryCACert:        strings.TrimSpace(v.GetString("CONTAINER_REGISTRY_CA_CERT")),
		LogsDir:                        strings.TrimSpace(v.GetString("LOGS_DIR")),
		TemporalAddress:                strings.TrimSpace(v.GetString("TEMPORAL_ADDRESS")),
		TemporalNamespace:              strings.TrimSpace(v.GetString("TEMPORAL_NAMESPACE")),
		TemporalEnableTLS:              v.GetBool("TEMPORAL_ENABLE_TLS"),
		TemporalAPIKey:                 strings.TrimSpace(v.GetString("TEMPORAL_API_KEY")),
		TemporalExternal:               v.GetBool("TEMPORAL_EXTERNAL"),
		TemporalTaskQueue:              strings.TrimSpace(v.GetString("TEMPORAL_TASK_QUEUE")),
		MaxMemory:                      v.GetInt64("MAX_MEMORY"),
		MaxUploadSize:                  v.GetInt64("MAX_UPLOAD_SIZE"),
		SessionOn:                      v.GetBool("SESSION_ON"),

		PostgresDSN:           strings.TrimSpace(v.GetString("POSTGRES_DB")),
		EncryptionKey:         strings.TrimSpace(v.GetString("OLAKE_SECRET_KEY")),
		OlakePostgresUser:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_USER")),
		OlakePostgresPassword: strings.TrimSpace(v.GetString("OLAKE_POSTGRES_PASSWORD")),
		OlakePostgresHost:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_HOST")),
		OlakePostgresPort:     strings.TrimSpace(v.GetString("OLAKE_POSTGRES_PORT")),
		OlakePostgresDBName:   strings.TrimSpace(v.GetString("OLAKE_POSTGRES_DBNAME")),
		OlakePostgresSSLMode:  strings.TrimSpace(v.GetString("OLAKE_POSTGRES_SSLMODE")),

		MetricsEnabled: v.GetBool("METRICS_ENABLED"),

		EnableOptimization:         v.GetBool("ENABLE_OPTIMIZATION"),
		OptimizationGroup:          strings.TrimSpace(v.GetString("OPTIMIZATION_GROUP")),
		OptimizationBaseURL:        strings.TrimSpace(v.GetString("OPTIMIZATION_BASE_URL")),
		OptimizationUsername:       strings.TrimSpace(v.GetString("USERNAME")),
		OptimizationPassword:       strings.TrimSpace(v.GetString("PASSWORD")),
		OptimizationRequestTimeout: v.GetDuration("OPTIMIZATION_REQUEST_TIMEOUT"),

		GitOpsMode:             strings.ToLower(strings.TrimSpace(v.GetString("GITOPS_MODE"))),
		PodNamespace:           strings.TrimSpace(v.GetString("POD_NAMESPACE")),
		GitOpsFileDir:          strings.TrimSpace(v.GetString("OLAKE_GITOPS_FILE_DIR")),
		GitOpsStateFile:        strings.TrimSpace(v.GetString("OLAKE_GITOPS_STATE_FILE")),
		OlakeStorageMode:       strings.TrimSpace(v.GetString("OLAKE_STORAGE_MODE")),
		OlakeS3Bucket:          strings.TrimSpace(v.GetString("OLAKE_S3_BUCKET")),
		OlakeS3Region:          strings.TrimSpace(v.GetString("OLAKE_S3_REGION")),
		OlakeS3Prefix:          strings.TrimSpace(v.GetString("OLAKE_S3_PREFIX")),
		OlakeS3AccessKeyID:     strings.TrimSpace(v.GetString("OLAKE_S3_ACCESS_KEY_ID")),
		OlakeS3SecretAccessKey: strings.TrimSpace(v.GetString("OLAKE_S3_SECRET_ACCESS_KEY")),
		OlakeS3Endpoint:        strings.TrimSpace(v.GetString("OLAKE_S3_ENDPOINT")),

		GitOpsSecretsRefreshInterval:   v.GetDuration("GITOPS_SECRETS_REFRESH_INTERVAL"),
		GitOpsSecretsAWSRegion:         strings.TrimSpace(v.GetString("GITOPS_SECRETS_AWS_REGION")),
		GitOpsSecretsAWSAccessKeyID:    strings.TrimSpace(v.GetString("GITOPS_SECRETS_AWS_ACCESS_KEY_ID")),
		GitOpsSecretsAWSSecretKey:      strings.TrimSpace(v.GetString("GITOPS_SECRETS_AWS_SECRET_ACCESS_KEY")),
		GitOpsSecretsGCPProject:        strings.TrimSpace(v.GetString("GITOPS_SECRETS_GCP_PROJECT")),
		GitOpsSecretsAzureTenantID:     strings.TrimSpace(v.GetString("GITOPS_SECRETS_AZURE_TENANT_ID")),
		GitOpsSecretsAzureClientID:     strings.TrimSpace(v.GetString("GITOPS_SECRETS_AZURE_CLIENT_ID")),
		GitOpsSecretsAzureClientSecret: strings.TrimSpace(v.GetString("GITOPS_SECRETS_AZURE_CLIENT_SECRET")),
	}
}
