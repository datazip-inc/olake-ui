package models

// Common fields for source/destination config
type ConnectorConfig struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}

// LoginRequest represents the expected JSON structure for login requests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Spec request for getting specs
type SpecRequest struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Catalog string `json:"catalog"`
}

// Test connection requests
type SourceTestConnectionRequest struct {
	ConnectorConfig
	SourceID int `json:"source_id"`
}

type DestinationTestConnectionRequest struct {
	ConnectorConfig
}

// Create/Update source and destination requests
type CreateSourceRequest struct {
	ConnectorConfig
}

type UpdateSourceRequest struct {
	ConnectorConfig
}

type CreateDestinationRequest struct {
	ConnectorConfig
}

type UpdateDestinationRequest struct {
	ConnectorConfig
}

// Job source and destination configurations
type JobSourceConfig = ConnectorConfig
type JobDestinationConfig = ConnectorConfig

// Create and update job requests
type CreateJobRequest struct {
	Name          string               `json:"name"`
	Source        JobSourceConfig      `json:"source"`
	Destination   JobDestinationConfig `json:"destination"`
	Frequency     string               `json:"frequency"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)"`
	Activate      bool                 `json:"activate,omitempty"`
}

type UpdateJobRequest struct {
	Name          string               `json:"name"`
	Source        JobSourceConfig      `json:"source"`
	Destination   JobDestinationConfig `json:"destination"`
	Frequency     string               `json:"frequency"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)"`
	Activate      bool                 `json:"activate,omitempty"`
}
