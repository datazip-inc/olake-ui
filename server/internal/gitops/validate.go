package gitops

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

var validate = validator.New()

func init() {
	// DTOs use binding:"required" (Gin); read those tags here instead of validate:"required".
	validate.SetTagName("binding")
}

// GitOpsJobConfig is the JSON shape stored in Job CR spec.config.
type GitOpsJobConfig struct {
	Name             string                `json:"name" binding:"required"`
	Source           string                `json:"source" binding:"required"`      // OLake source name or numeric ID
	Destination      string                `json:"destination" binding:"required"` // OLake destination name or numeric ID
	Frequency        string                `json:"frequency" binding:"required"`
	Activate         bool                  `json:"activate"`
	AdvancedSettings *dto.AdvancedSettings `json:"advanced_settings,omitempty"`
}

func ParseAndValidateSource(config string) (*dto.CreateSourceRequest, error) {
	var req dto.CreateSourceRequest
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		return nil, Permanent(fmt.Errorf("invalid spec.config JSON: %w", err))
	}
	if err := validate.Struct(req); err != nil {
		return nil, Permanent(fmt.Errorf("validation failed: %w", err))
	}
	if err := dto.ValidateSourceType(req.Type); err != nil {
		return nil, Permanent(err)
	}
	return &req, nil
}

func ParseAndValidateDestination(config string) (*dto.CreateDestinationRequest, error) {
	var req dto.CreateDestinationRequest
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		return nil, Permanent(fmt.Errorf("invalid spec.config JSON: %w", err))
	}
	if err := validate.Struct(req); err != nil {
		return nil, Permanent(fmt.Errorf("validation failed: %w", err))
	}
	if err := dto.ValidateDestinationType(req.Type); err != nil {
		return nil, Permanent(err)
	}
	return &req, nil
}

func ParseAndValidateJobConfig(config string) (*GitOpsJobConfig, error) {
	var req GitOpsJobConfig
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		return nil, Permanent(fmt.Errorf("invalid spec.config JSON: %w", err))
	}
	if err := validate.Struct(req); err != nil {
		return nil, Permanent(fmt.Errorf("validation failed: %w", err))
	}
	return &req, nil
}
