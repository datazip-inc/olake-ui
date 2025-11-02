package dto

// Common fields for source/destination config
// source and destination are driver in olake cli
type DriverConfig struct {
	ID      *int   `json:"id,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Source  string `json:"source_type"`
	Config  string `json:"config" orm:"type(jsonb)"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SpecRequest struct {
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
}

// check unique job name request
type CheckUniqueJobNameRequest struct {
	JobName string `json:"job_name"`
}

// Test connection requests
type SourceTestConnectionRequest struct {
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}
type StreamsRequest struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
	JobID   int    `json:"job_id" validate:"required"`
	JobName string `json:"job_name" validate:"required"`
}

// TODO: frontend needs to send only version no need for source type and version
type DestinationTestConnectionRequest struct {
	Type          string `json:"type" validate:"required"`
	Version       string `json:"version" validate:"required"`
	Config        string `json:"config" validate:"required"`
	SourceType    string `json:"source_type"`
	SourceVersion string `json:"source_version"`
}

type CreateSourceRequest struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}

type UpdateSourceRequest struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}

type CreateDestinationRequest struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}

type UpdateDestinationRequest struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}

type CreateJobRequest struct {
	Name          string        `json:"name" validate:"required"`
	Source        *DriverConfig `json:"source" validate:"required"`
	Destination   *DriverConfig `json:"destination" validate:"required"`
	Frequency     string        `json:"frequency" validate:"required"`
	StreamsConfig string        `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	Activate      bool          `json:"activate,omitempty"`
}

type UpdateJobRequest struct {
	Name              string        `json:"name" validate:"required"`
	Source            *DriverConfig `json:"source" validate:"required"`
	Destination       *DriverConfig `json:"destination" validate:"required"`
	Frequency         string        `json:"frequency" validate:"required"`
	StreamsConfig     string        `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	DifferenceStreams string        `json:"difference_streams,omitempty"`
	Activate          bool          `json:"activate,omitempty"`
}

type StreamDifferenceRequest struct {
	UpdatedStreamsConfig string `json:"updated_streams_config" validate:"required"`
}

type JobTaskRequest struct {
	FilePath string `json:"file_path" validate:"required"`
}

type JobStatusRequest struct {
	Activate bool `json:"activate"`
}
