package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// GetProjectSettingsByProjectID fetches the settings row for a project ID.
func (db *Database) GetProjectSettingsByProjectID(projectID string) (*models.ProjectSettings, error) {
	settings := &models.ProjectSettings{}
	err := db.conn.
		Where("project_id = ?", projectID).
		First(settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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

	row := &models.ProjectSettings{
		ID:              0,
		ProjectID:       settings.ProjectID,
		WebhookAlertURL: settings.WebhookAlertURL,
	}
	if err := db.conn.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "project_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"webhook_alert_url": settings.WebhookAlertURL,
			}),
		}).
		Create(row).Error; err != nil {
		return fmt.Errorf("failed to upsert project settings project_id[%s]: %s", settings.ProjectID, err)
	}
	return nil
}
