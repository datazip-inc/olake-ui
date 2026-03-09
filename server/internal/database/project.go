package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// GetProjectSettingsByProjectID fetches the settings row for a project ID.
func (db *Database) GetProjectSettingsByProjectID(projectID string) (*models.ProjectSettings, error) {
	settings := &models.ProjectSettings{}
	err := db.conn.
		Where("project_id = ?", projectID).
		First(settings).Error
	if err == gorm.ErrRecordNotFound {
		return &models.ProjectSettings{ProjectID: projectID, ID: 0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project settings project_id[%s]: %s", projectID, err)
	}

	return settings, nil
}

// UpsertProjectSettingsModel upserts (inserts or updates) project settings.
func (db *Database) UpsertProjectSettingsModel(settings *models.ProjectSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}

	existing := &models.ProjectSettings{}
	err := db.conn.Where("project_id = ?", settings.ProjectID).First(existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to lookup project settings project_id[%s]: %s", settings.ProjectID, err)
	}
	if err == gorm.ErrRecordNotFound {
		settings.ID = 0
		if err := db.conn.Create(settings).Error; err != nil {
			return fmt.Errorf("failed to insert project settings project_id[%s]: %s", settings.ProjectID, err)
		}
		return nil
	}

	settings.ID = existing.ID
	if err := db.conn.Model(&models.ProjectSettings{}).
		Where("id = ?", settings.ID).
		Updates(map[string]any{
			"project_id":        settings.ProjectID,
			"webhook_alert_url": settings.WebhookAlertURL,
		}).Error; err != nil {
		return fmt.Errorf("failed to update project settings project_id[%s]: %s", settings.ProjectID, err)
	}
	return nil
}
