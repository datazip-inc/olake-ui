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

type CheckUniqueJobNameRequest struct {
	JobName string `json:"job_name"`
}

// Common fields for source/destination config
type ConnectorConfig struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
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

type SourceTestConnectionRequest struct {
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
}

type StreamsRequest struct {
	ConnectorConfig
	JobID   int    `json:"job_id" validate:"required"`
	JobName string `json:"job_name"`
}

type DestinationTestConnectionRequest struct {
	Type          string `json:"type" validate:"required"`
	Version       string `json:"version" validate:"required"`
	Config        string `json:"config" orm:"type(jsonb)" validate:"required"`
	SourceType    string `json:"source_type"`
	SourceVersion string `json:"source_version"`
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
	Source        JobSourceConfig      `json:"source" validate:"required"`
	Destination   JobDestinationConfig `json:"destination" validate:"required"`
	Frequency     string               `json:"frequency" validate:"required"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	Activate      bool                 `json:"activate,omitempty"`
}

type UpdateJobRequest struct {
	Name          string               `json:"name" validate:"required"`
	Source        JobSourceConfig      `json:"source" validate:"required"`
	Destination   JobDestinationConfig `json:"destination" validate:"required"`
	Frequency     string               `json:"frequency" validate:"required"`
	StreamsConfig string               `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	Activate      bool                 `json:"activate,omitempty"`
}

type JobTaskRequest struct {
	FilePath string `json:"file_path" validate:"required"`
}
