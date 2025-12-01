package services

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

func (s *ETLService) GetProjectSettings(ctx context.Context, projectID string) (dto.ProjectSettingsResponse, error) {
	settings, err := s.db.GetProjectSettingsByProjectID(projectID)
	if err != nil {
		return dto.ProjectSettingsResponse{}, fmt.Errorf("failed to get project settings: %s", err)
	}

	if settings == nil {
		return dto.ProjectSettingsResponse{}, fmt.Errorf("project settings not found: %s", err)
	}

	return dto.ProjectSettingsResponse{
		ID:              settings.ID,
		ProjectID:       settings.ProjectID,
		WebhookAlertURL: settings.WebhookAlertURL,
	}, nil
}

func (s *ETLService) UpdateProjectSettings(ctx context.Context, projectID string, req dto.UpdateProjectSettingsRequest) error {
	projectSettings := &models.ProjectSettings{
		ID:              req.ID,
		ProjectID:       req.ProjectID,
		WebhookAlertURL: req.WebhookAlertURL,
	}

	if err := s.db.UpsertProjectSettingsModel(projectSettings); err != nil {
		return fmt.Errorf("failed to update project settings: %s", err)
	}

	return nil
}
