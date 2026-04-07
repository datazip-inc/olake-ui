package dto

// Common fields for source/destination config
// source and destination are driver in olake cli
type DriverConfig struct {
	ID      *int   `json:"id,omitempty" example:"1"`
	Name    string `json:"name" example:"my-postgres-source"`
	Type    string `json:"type" example:"postgres"`
	Version string `json:"version" example:"v0.2.7"`
	Config  string `json:"config,omitempty" orm:"type(jsonb)" example:"{\"host\":\"localhost\",\"port\":5432}"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password"`
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
}

type UpdateUserRequest struct {
	Username string `json:"username" example:"admin"`
	Email    string `json:"email" example:"admin@example.com"`
}

type SpecRequest struct {
	// enum: postgres,mongodb,mysql,mssql,db2,s3,kafka,iceberg
	Type    string `json:"type" binding:"required" example:"postgres"`
	Version string `json:"version" binding:"required" example:"v0.2.7"`
}

// check unique job name request
type CheckUniqueNameRequest struct {
	Name       string `json:"name" binding:"required" example:"my-sync-job"`
	EntityType string `json:"entity_type" binding:"required,oneof=job source destination" example:"job"`
}

// Test connection requests
type SourceTestConnectionRequest struct {
	Type    string `json:"type" binding:"required" example:"postgres"`
	Version string `json:"version" binding:"required" example:"v0.2.7"`
	Config  string `json:"config" orm:"type(jsonb)" binding:"required" example:"{\"host\":\"localhost\",\"port\":5432,\"database\":\"mydb\",\"user\":\"postgres\",\"password\":\"secret\"}"`
}
type StreamsRequest struct {
	Name               string `json:"name" binding:"required" example:"my-postgres-source"`
	Type               string `json:"type" binding:"required" example:"postgres"`
	Version            string `json:"version" binding:"required" example:"v0.2.7"`
	Config             string `json:"config" orm:"type(jsonb)" binding:"required" example:"{\"host\":\"localhost\",\"port\":5432}"`
	MaxDiscoverThreads *int   `json:"max_discover_threads,omitempty" example:"50"`
	JobID              int    `json:"job_id" binding:"required" example:"1"`
	JobName            string `json:"job_name" binding:"required" example:"my-sync-job"`
}

// TODO: frontend needs to send only version no need for source version
type DestinationTestConnectionRequest struct {
	Type          string `json:"type" binding:"required" example:"iceberg"`
	Version       string `json:"version" binding:"required" example:"v0.2.7"`
	Config        string `json:"config" binding:"required" example:"{\"catalog_type\":\"glue\",\"warehouse\":\"s3://my-bucket/warehouse\"}"`
	SourceType    string `json:"source_type" example:"postgres"`
	SourceVersion string `json:"source_version" example:"v0.2.7"`
}

type CreateSourceRequest struct {
	Name    string `json:"name" binding:"required" example:"my-postgres-source"`
	Type    string `json:"type" binding:"required" example:"postgres"`
	Version string `json:"version" binding:"required" example:"v0.2.7"`
	Config  string `json:"config" orm:"type(jsonb)" binding:"required" example:"{\"host\":\"localhost\",\"port\":5432,\"database\":\"mydb\",\"user\":\"postgres\",\"password\":\"secret\"}"`
}

type UpdateSourceRequest struct {
	Name    string `json:"name" binding:"required" example:"my-postgres-source-updated"`
	Type    string `json:"type" binding:"required" example:"postgres"`
	Version string `json:"version" binding:"required" example:"v0.2.7"`
	Config  string `json:"config" orm:"type(jsonb)" binding:"required" example:"{\"host\":\"localhost\",\"port\":5432,\"database\":\"mydb\",\"user\":\"postgres\",\"password\":\"newsecret\"}"`
}

type CreateDestinationRequest struct {
	Name    string `json:"name" binding:"required" example:"my-iceberg-destination"`
	Type    string `json:"type" binding:"required" example:"iceberg"`
	Version string `json:"version" binding:"required" example:"v0.2.7"`
	Config  string `json:"config" orm:"type(jsonb)" binding:"required" example:"{\"catalog_type\":\"glue\",\"warehouse\":\"s3://my-bucket/warehouse\"}"`
}

type UpdateDestinationRequest struct {
	Name    string `json:"name" binding:"required" example:"my-iceberg-destination-updated"`
	Type    string `json:"type" binding:"required" example:"iceberg"`
	Version string `json:"version" binding:"required" example:"v0.2.7"`
	Config  string `json:"config" orm:"type(jsonb)" binding:"required" example:"{\"catalog_type\":\"glue\",\"warehouse\":\"s3://my-bucket/warehouse-v2\"}"`
}

type AdvancedSettings struct {
	MaxDiscoverThreads *int `json:"max_discover_threads,omitempty" example:"50"`
}

type CreateJobRequest struct {
	Name             string            `json:"name" binding:"required" example:"my-sync-job"`
	Source           *DriverConfig     `json:"source" binding:"required"`
	Destination      *DriverConfig     `json:"destination" binding:"required"`
	Frequency        string            `json:"frequency" binding:"required" example:"0 */6 * * *"`
	StreamsConfig    string            `json:"streams_config" orm:"type(jsonb)" binding:"required"`
	Activate         bool              `json:"activate,omitempty" example:"true"`
	AdvancedSettings *AdvancedSettings `json:"advanced_settings,omitempty"`
}

type UpdateJobRequest struct {
	Name              string            `json:"name" binding:"required" example:"my-sync-job-updated"`
	Source            *DriverConfig     `json:"source" binding:"required"`
	Destination       *DriverConfig     `json:"destination" binding:"required"`
	Frequency         string            `json:"frequency" binding:"required" example:"0 */12 * * *"`
	StreamsConfig     string            `json:"streams_config" orm:"type(jsonb)" binding:"required"`
	DifferenceStreams string            `json:"difference_streams,omitempty" example:"[]"`
	Activate          bool              `json:"activate,omitempty" example:"true"`
	AdvancedSettings  *AdvancedSettings `json:"advanced_settings,omitempty"`
}

type StreamDifferenceRequest struct {
	UpdatedStreamsConfig string `json:"updated_streams_config" binding:"required"`
}

type JobTaskRequest struct {
	FilePath string `json:"file_path" binding:"required" example:"sync-123-2-2026-01-19T13:45:09Z"`
}

type JobStatusRequest struct {
	Activate bool `json:"activate" example:"true"`
}

type UpsertProjectSettingsRequest struct {
	ID              int    `json:"id" example:"1"`
	ProjectID       string `json:"project_id" binding:"required" example:"project-123"`
	WebhookAlertURL string `json:"webhook_alert_url" example:"https://hooks.slack.com/services/xxx/yyy/zzz"`
}

type UpdateSyncTelemetryRequest struct {
	JobID       int    `json:"job_id"`
	WorkflowID  string `json:"workflow_id"`
	Event       string `json:"event"`
	Environment string `json:"environment"`
}

type UpdateStateFileRequest struct {
	StateFile string `json:"state_file" binding:"required"`
}

type CatalogRequest struct {
	Name                    string            `json:"name" binding:"required"`
	Type                    string            `json:"type" binding:"required"`
	OptimizerGroup          string            `json:"optimizerGroup" binding:"required"`
	OptimizeTableFormatList []string          `json:"tableFormatList"`
	StorageConfig           map[string]string `json:"storageConfig" binding:"required"`
	AuthConfig              map[string]string `json:"authConfig" binding:"required"`
	Properties              map[string]string `json:"properties" binding:"required"`
	TableProperties         map[string]string `json:"tableProperties" binding:"required"`
}

// LogInfo represents the log information from terminal execution
type LogInfo struct {
	LogStatus string   `json:"logStatus"` // "Finished", "Failed", etc.
	Logs      []string `json:"logs"`
}

type SetTablePropertiesRequest struct {
	Catalog    string            `json:"catalog" binding:"required"`
	Database   string            `json:"database" binding:"required"`
	Table      string            `json:"table" binding:"required"`
	Properties map[string]string `json:"properties" binding:"required"`
}

// TerminalExecuteRequest represents the request body for terminal SQL execution
type TerminalExecuteRequest struct {
	SQL string `json:"sql" binding:"required"`
}

type SQLInput struct {
	MinorCron              *string `json:"minor_cron"`
	MajorCron              *string `json:"major_cron"`
	FullCron               *string `json:"full_cron"`
	TargetFileSize         *int64  `json:"target_file_size"`
	EnabledForOptimization *string `json:"enabled_for_optimization"`
}

type CatalogRequest struct {
	Name                    string            `json:"name"`
	Type                    string            `json:"type"`
	OptimizerGroup          string            `json:"optimizerGroup"`
	OptimizeTableFormatList []string          `json:"tableFormatList"`
	StorageConfig           map[string]string `json:"storageConfig"`
	AuthConfig              map[string]string `json:"authConfig"`
	Properties              map[string]string `json:"properties"`
	TableProperties         map[string]string `json:"tableProperties"`
}

// LogInfo represents the log information from terminal execution
type LogInfo struct {
	LogStatus string   `json:"logStatus"` // "Finished", "Failed", etc.
	Logs      []string `json:"logs"`
}

type SetTablePropertiesRequest struct {
	Catalog    string            `json:"catalog"`
	Database   string            `json:"database"`
	Table      string            `json:"table"`
	Properties map[string]string `json:"properties"`
}

// TerminalExecuteRequest represents the request body for terminal SQL execution
type TerminalExecuteRequest struct {
	SQL string `json:"sql"`
}

type SQLInput struct {
	MinorCron              *string `json:"minor_cron"`
	MajorCron              *string `json:"major_cron"`
	FullCron               *string `json:"full_cron"`
	TargetFileSize         *int64  `json:"target_file_size"`
	EnabledForOptimization *string `json:"enabled_for_optimization"`
}
