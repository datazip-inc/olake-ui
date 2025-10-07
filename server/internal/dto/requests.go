package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidateStruct validates any struct that has `validate` tags.
func Validate(s interface{}) error {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("invalid validation error: %v", err)
		}

		// collect all validation errors into a single message
		var errorMessages string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages += fmt.Sprintf("Field '%s' failed validation rule '%s'; ", err.Field(), err.Tag())
		}
		return fmt.Errorf("validation failed: %s", errorMessages)
	}
	return nil
}

// Common fields for source/destination config
type ConnectorConfig struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Source  string `json:"source_type" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SpecRequest struct {
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Catalog string `json:"catalog"`
}

// check unique job name request
type CheckUniqueJobNameRequest struct {
	JobName string `json:"job_name"`
}

// Test connection requests
type SourceTestConnectionRequest struct {
	ConnectorConfig
	SourceID int `json:"source_id"`
}
type StreamsRequest struct {
	ConnectorConfig
	JobID   int    `json:"job_id"`
	JobName string `json:"job_name"`
}
type DestinationTestConnectionRequest struct {
	ConnectorConfig
}

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

type CreateJobRequest struct {
	Name          string               `json:"name" validate:"required"`
	Source        JobSourceConfig      `json:"source" validate:"required,dive"`
	Destination   JobDestinationConfig `json:"destination" validate:"required,dive"`
	Frequency     string               `json:"frequency" validate:"required"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	Activate      bool                 `json:"activate,omitempty"`
}

type UpdateJobRequest struct {
	Name          string               `json:"name" validate:"required"`
	Source        JobSourceConfig      `json:"source" validate:"required,dive"`
	Destination   JobDestinationConfig `json:"destination" validate:"required,dive"`
	Frequency     string               `json:"frequency" validate:"required"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	Activate      bool                 `json:"activate,omitempty"`
}

type JobTaskRequest struct {
	FilePath string `json:"file_path" validate:"required"`
}
