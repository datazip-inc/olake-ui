package gitops

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

var validate = validator.New()

func init() {
	validate.SetTagName("binding")
}

// JobConfig is the JSON shape stored in Job ConfigMap data.config.
// Source/Destination are name or numeric ID refs (not *DriverConfig like the HTTP API).
type JobConfig struct {
	dto.JobMetadata
	Source      string `json:"source" binding:"required"`      // OLake source name or numeric ID
	Destination string `json:"destination" binding:"required"` // OLake destination name or numeric ID
}

func ParseAndValidateSource(config string) (*dto.CreateSourceRequest, error) {
	var req dto.CreateSourceRequest
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		return nil, NonRetryableError(fmt.Errorf("invalid spec.config JSON: %w", err))
	}
	if err := validate.Struct(req); err != nil {
		return nil, NonRetryableError(fmt.Errorf("validation failed: %w", err))
	}
	if err := dto.ValidateSourceType(req.Type); err != nil {
		return nil, NonRetryableError(err)
	}
	return &req, nil
}

func ParseAndValidateDestination(config string) (*dto.CreateDestinationRequest, error) {
	var req dto.CreateDestinationRequest
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		return nil, NonRetryableError(fmt.Errorf("invalid spec.config JSON: %w", err))
	}
	if err := validate.Struct(req); err != nil {
		return nil, NonRetryableError(fmt.Errorf("validation failed: %w", err))
	}
	if err := dto.ValidateDestinationType(req.Type); err != nil {
		return nil, NonRetryableError(err)
	}
	return &req, nil
}

func ParseAndValidateJobConfig(config string) (*JobConfig, error) {
	var req JobConfig
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		return nil, NonRetryableError(fmt.Errorf("invalid spec.config JSON: %w", err))
	}
	if err := validate.Struct(req); err != nil {
		return nil, NonRetryableError(fmt.Errorf("validation failed: %w", err))
	}
	return &req, nil
}
