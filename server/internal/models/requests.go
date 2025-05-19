package models

// LoginRequest represents the expected JSON structure for login requests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type SpecRequest struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Catalog string `json:"catalog"`
}
type SourceTestConnectionRequest struct {
	Type     string `json:"type"`
	Version  string `json:"version"`
	Config   string `json:"config" orm:"type(jsonb)"`
	SourceID int    `json:"source_id"`
}
type DestinationTestConnectionRequest struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type CreateSourceRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type UpdateSourceRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type CreateDestinationRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type UpdateDestinationRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}

// Job source configuration
type JobSourceConfig struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Config  string `json:"config"`
	Version string `json:"version"`
}

// Job destination configuration
type JobDestinationConfig struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Config  string `json:"config"`
	Version string `json:"version"`
}

// CreateJobRequest represents the request body for creating a job
type CreateJobRequest struct {
	Name          string               `json:"name"`
	Source        JobSourceConfig      `json:"source"`
	Destination   JobDestinationConfig `json:"destination"`
	Frequency     string               `json:"frequency"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)"`
	Activate      bool                 `json:"activate,omitempty"`
}

// UpdateJobRequest represents the request body for updating a job
type UpdateJobRequest struct {
	Name          string               `json:"name"`
	Source        JobSourceConfig      `json:"source"`
	Destination   JobDestinationConfig `json:"destination"`
	Frequency     string               `json:"frequency"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)"`
	Activate      bool                 `json:"activate,omitempty"`
}
