package dto

type JSONResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Bad request"`
}

// Error400Response represents a 400 Bad Request error
type Error400Response struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Bad request"`
}

// Error401Response represents a 401 Unauthorized error
type Error401Response struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Authentication required"`
}

// Error404Response represents a 404 Not Found error
type Error404Response struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Resource not found"`
}

// Error409Response represents a 409 Conflict error
type Error409Response struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Resource already exists"`
}

// Error500Response represents a 500 Internal Server Error
type Error500Response struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Internal server error"`
}

type SpecResponse struct {
	Version string      `json:"version" example:"v0.2.7"`
	Type    string      `json:"type" example:"postgres"`
	Spec    interface{} `json:"spec" orm:"type(jsonb)" swaggertype:"object"`
}
type SpecOutput struct {
	Spec map[string]interface{} `json:"spec"`
}

type DeleteSourceResponse struct {
	Name string `json:"name" example:"my-postgres-source"`
}

type DeleteDestinationResponse struct {
	Name string `json:"name" example:"my-iceberg-destination"`
}

type JobStatus struct {
	Activate bool `json:"activate" example:"true"`
}

// JobNameCheckResponse
type CheckUniqueJobNameResponse struct {
	Unique bool `json:"unique" example:"true"`
}

// TestConnectionResponse
type TestConnectionResponse struct {
	ConnectionResult map[string]interface{}   `json:"connection_result" swaggertype:"object"`
	Logs             []map[string]interface{} `json:"logs" swaggertype:"array,object"`
}

type StreamDifferenceResponse struct {
	DifferenceStreams map[string]interface{} `json:"difference_streams" swaggertype:"object"`
}

type ClearDestinationStatusResponse struct {
	Running bool `json:"running" example:"false"`
}

type VersionsResponse struct {
	Version []string `json:"version" example:"v0.3.15,v0.3.14,v0.3.13,v0.3.12"`
}

// Job response
type JobResponse struct {
	ID               int               `json:"id" example:"1"`
	Name             string            `json:"name" example:"my-sync-job"`
	Source           DriverConfig      `json:"source"`
	Destination      DriverConfig      `json:"destination"`
	StreamsConfig    string            `json:"streams_config,omitempty"`
	Frequency        string            `json:"frequency" example:"0 */6 * * *"`
	LastRunTime      string            `json:"last_run_time,omitempty" example:"2024-01-09T12:00:00Z"`
	LastRunState     string            `json:"last_run_state,omitempty" example:"completed"`
	LastRunType      string            `json:"last_run_type,omitempty" example:"sync"` // "sync" | "clear-destination"
	CreatedAt        string            `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt        string            `json:"updated_at" example:"2024-01-09T12:00:00Z"`
	Activate         bool              `json:"activate" example:"true"`
	CreatedBy        string            `json:"created_by,omitempty" example:"admin"`
	UpdatedBy        string            `json:"updated_by,omitempty" example:"admin"`
	AdvancedSettings *AdvancedSettings `json:"advanced_settings,omitempty"`
}

type JobTask struct {
	Runtime   string `json:"runtime" example:"1m30s"`
	StartTime string `json:"start_time" example:"2024-01-09T12:00:00Z"`
	Status    string `json:"status" example:"completed"`
	FilePath  string `json:"file_path" example:"sync-123-2-2026-01-19T13:45:09Z"`
	JobType   string `json:"job_type" example:"sync"` // "sync" | "clear-destination"
}

type SourceDataItem struct {
	ID        int           `json:"id" example:"1"`
	Name      string        `json:"name" example:"my-postgres-source"`
	Type      string        `json:"type" example:"postgres"`
	Version   string        `json:"version" example:"v0.2.7"`
	Config    string        `json:"config" example:"{\"host\":\"localhost\",\"port\":5432}"`
	CreatedAt string        `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string        `json:"updated_at" example:"2024-01-09T12:00:00Z"`
	CreatedBy string        `json:"created_by" example:"admin"`
	UpdatedBy string        `json:"updated_by" example:"admin"`
	Jobs      []JobDataItem `json:"jobs"`
}

type DestinationDataItem struct {
	ID        int           `json:"id" example:"1"`
	Name      string        `json:"name" example:"my-iceberg-destination"`
	Type      string        `json:"type" example:"iceberg"`
	Version   string        `json:"version" example:"v0.2.7"`
	Config    string        `json:"config" example:"{\"catalog_type\":\"glue\",\"warehouse\":\"s3://bucket/warehouse\"}"`
	CreatedAt string        `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string        `json:"updated_at" example:"2024-01-09T12:00:00Z"`
	CreatedBy string        `json:"created_by" example:"admin"`
	UpdatedBy string        `json:"updated_by" example:"admin"`
	Jobs      []JobDataItem `json:"jobs"`
}

type JobDataItem struct {
	Name            string `json:"name" example:"my-sync-job"`
	ID              int    `json:"id" example:"1"`
	Activate        bool   `json:"activate" example:"true"`
	SourceName      string `json:"source_name,omitempty" example:"my-postgres-source"`
	SourceType      string `json:"source_type,omitempty" example:"postgres"`
	DestinationName string `json:"destination_name,omitempty" example:"my-iceberg-destination"`
	DestinationType string `json:"destination_type,omitempty" example:"iceberg"`
	LastRunTime     string `json:"last_run_time,omitempty" example:"2024-01-09T12:00:00Z"`
	LastRunState    string `json:"last_run_state,omitempty" example:"completed"`
}

type TaskLogsResponse struct {
	Logs         []map[string]interface{} `json:"logs" swaggertype:"array,object"`
	OlderCursor  int64                    `json:"older_cursor" example:"1704801600000"`
	NewerCursor  int64                    `json:"newer_cursor" example:"1704805200000"`
	HasMoreOlder bool                     `json:"has_more_older" example:"true"`
	HasMoreNewer bool                     `json:"has_more_newer" example:"false"`
}

type ProjectSettingsResponse struct {
	ID              int    `json:"id" example:"1"`
	ProjectID       string `json:"project_id" example:"project-123"`
	WebhookAlertURL string `json:"webhook_alert_url" example:"https://hooks.slack.com/services/xxx/yyy/zzz"`
}

type LoginResponse struct {
	Username string `json:"username" example:"admin"`
}

// ReleaseMetadataResponse represents a single release
type ReleaseMetadataResponse struct {
	Title       string   `json:"title,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Date        string   `json:"date"`
	Link        string   `json:"link"`
}

// ReleaseTypeData represents releases for a specific type
type ReleaseTypeData struct {
	CurrentVersion string                     `json:"current_version,omitempty"`
	Releases       []*ReleaseMetadataResponse `json:"releases"`
}

// AllReleasesResponse represents the complete response
type ReleasesResponse struct {
	OlakeUIWorker *ReleaseTypeData `json:"olake_ui_worker"`
	OlakeHelm     *ReleaseTypeData `json:"olake_helm"`
	Olake         *ReleaseTypeData `json:"olake"`
	Features      *ReleaseTypeData `json:"features"`
}

type TelemetryIDResponse struct {
	TelemetryUserID string `json:"user_id" example:"1234567890abcdef1234567890abcdef"`
	OlakeUIVersion  string `json:"version" example:"v0.2.5"`
}
