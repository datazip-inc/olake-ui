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
// source and destiantion are driver in olake cli
type driverConfig struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Version string `json:"version" validate:"required"`
	Source  string `json:"source_type"`
	Config  string `json:"config" orm:"type(jsonb)" validate:"required"`
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

// Job source and destination configurations
type JobSourceConfig = driverConfig
type JobDestinationConfig = driverConfig

type CreateJobRequest struct {
	Name          string                `json:"name" validate:"required"`
	Source        *JobSourceConfig      `json:"source" validate:"required"`
	Destination   *JobDestinationConfig `json:"destination" validate:"required"`
	Frequency     string                `json:"frequency" validate:"required"`
	StreamsConfig string                `json:"streams_config" orm:"type(jsonb)" validate:"required"`
	Activate      bool                  `json:"activate,omitempty"`
}

type UpdateJobRequest = CreateJobRequest

type JobTaskRequest struct {
	FilePath string `json:"file_path" validate:"required"`
}
type JobStatusRequest struct {
	Activate bool `json:"activate" validate:"required"`
}
