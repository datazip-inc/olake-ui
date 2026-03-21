package models

// CatalogRequest represents the request to create or update a catalog
type CatalogRequest struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	OptimizerGroup  string            `json:"optimizerGroup,omitempty"`
	TableFormatList []string          `json:"tableFormatList"`
	StorageConfig   map[string]string `json:"storageConfig"`
	AuthConfig      map[string]string `json:"authConfig"`
	Properties      map[string]string `json:"properties"`
	TableProperties map[string]string `json:"tableProperties"`
}

// CatalogResponse represents the response from catalog operations
type CatalogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CatalogsResponse represents catalogs with their databases (no table details)
type CatalogsResponse struct {
	Catalogs []CatalogWithDatabases `json:"catalogs"`
}

type CatalogWithDatabases struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Databases    []string `json:"databases"`
	CreatedAt    string   `json:"created_at"`
	OLakeCreated bool     `json:"olake_created"`
}

// TablesResponse represents tables with full details for a specific catalog/database
type TablesResponse struct {
	Catalog  string      `json:"catalog"`
	Database string      `json:"database"`
	Tables   []TableInfo `json:"tables"`
}

type TableInfo struct {
	Name        string            `json:"name"`
	TotalSize   string            `json:"totalSize,omitempty"`
	ByOLake     bool              `json:"byOLake"`
	Major       *OptimizationInfo `json:"major"`
	Minor       *OptimizationInfo `json:"minor"`
	Full        *OptimizationInfo `json:"full"`
	Enabled     bool              `json:"enabled"`
	HealthScore int               `json:"healthScore,omitempty"`
}

type OptimizationInfo struct {
	LastRun string `json:"last-run,omitempty"`
	Status  string `json:"status,omitempty"`
}

// SetTablePropertiesRequest represents the request to set table properties
type SetTablePropertiesRequest struct {
	Catalog    string            `json:"catalog"`
	Database   string            `json:"database"`
	Table      string            `json:"table"`
	Properties map[string]string `json:"properties"`
}

// CompactionCronConfigRequest represents the request to set compaction cron configuration
// TriggerInterval values are in milliseconds. Use "never" to disable a specific compaction type.
type CompactionCronConfigRequest struct {
	MinorTriggerInterval string `json:"minorTriggerInterval"`
	MajorTriggerInterval string `json:"majorTriggerInterval"`
	FullTriggerInterval  string `json:"fullTriggerInterval"`
	TargetFileSize       string `json:"targetFileSize"`
}

// CompactionCronConfigResponse represents the response from setting cron configuration
type CompactionCronConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Note: not importing it from "github.com/datazip-inc/olake", as it will bring 
// unnecessary dependencies in go.mod

// CatalogType represents supported Iceberg catalog implementations
type CatalogType string

const (
	// GlueCatalog is the AWS Glue catalog implementation
	GlueCatalog CatalogType = "glue"
	// JDBCCatalog is the JDBC catalog implementation
	JDBCCatalog CatalogType = "jdbc"
	// HiveCatalog is the Hive catalog implementation
	HiveCatalog CatalogType = "hive"
	// RestCatalog is the REST catalog implementation
	RestCatalog CatalogType = "rest"
)

type Config struct {
	// S3-compatible Storage Configuration
	Region             string `json:"aws_region,omitempty"`
	AccessKey          string `json:"aws_access_key,omitempty"`
	SecretKey          string `json:"aws_secret_key,omitempty"`
	SessionToken       string `json:"aws_session_token,omitempty"`
	ProfileName        string `json:"aws_profile,omitempty"`
	NoIdentifierFields bool   `json:"no_identifier_fields"` // Needed to set true for Databricks Unity Catalog as it doesn't support identifier fields

	// S3 endpoint for custom S3-compatible services (like MinIO)
	S3Endpoint  string `json:"s3_endpoint,omitempty"`
	S3UseSSL    bool   `json:"s3_use_ssl,omitempty"`    // Use HTTPS if true
	S3PathStyle bool   `json:"s3_path_style,omitempty"` // Use path-style instead of virtual-hosted-style https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html

	// Catalog Configuration
	CatalogType CatalogType `json:"catalog_type,omitempty"`
	CatalogName string      `json:"catalog_name,omitempty"`

	// Glue catalog optional overrides
	UseGlueAdditionalConfig bool   `json:"glue_additional_config,omitempty"`
	GlueEndpoint            string `json:"glue_endpoint,omitempty"`
	GlueAccessKey           string `json:"glue_access_key,omitempty"`
	GlueSecretKey           string `json:"glue_secret_key,omitempty"`
	GlueRegion              string `json:"glue_region,omitempty"`
	GlueCatalogID           string `json:"glue_catalog_id,omitempty"`

	// JDBC specific configuration
	JDBCUrl      string `json:"jdbc_url,omitempty"`
	JDBCUsername string `json:"jdbc_username,omitempty"`
	JDBCPassword string `json:"jdbc_password,omitempty"`

	// Hive specific configuration
	HiveURI         string `json:"hive_uri,omitempty"`
	HiveClients     int    `json:"hive_clients,omitempty"`
	HiveSaslEnabled bool   `json:"hive_sasl_enabled,omitempty"`

	// Iceberg Configuration
	IcebergDatabase string `json:"iceberg_db,omitempty"`
	IcebergS3Path   string `json:"iceberg_s3_path"`                // e.g. s3://bucket/path
	JarPath         string `json:"sink_jar_path,omitempty"`        // Path to the Iceberg sink JAR
	ServerHost      string `json:"sink_rpc_server_host,omitempty"` // gRPC server host

	// Rest Catalog Configuration
	RestCatalogURL    string `json:"rest_catalog_url,omitempty"`
	RestSigningName   string `json:"rest_signing_name,omitempty"`
	RestSigningRegion string `json:"rest_signing_region,omitempty"`
	RestSigningV4     bool   `json:"rest_signing_v_4,omitempty"`
	RestToken         string `json:"token,omitempty"`
	RestOAuthURI      string `json:"oauth2_uri,omitempty"`
	RestAuthType      string `json:"rest_auth_type,omitempty"`
	RestScope         string `json:"scope,omitempty"`
	RestCredential    string `json:"credential,omitempty"`

	UseArrowWrites bool `json:"arrow_writes,omitempty"`
}
