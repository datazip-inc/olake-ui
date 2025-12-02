package database

import (
	"fmt"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// GetProjectSettingsByProjectID fetches the settings row for a project ID.
func (db *Database) GetProjectSettingsByProjectID(projectID string) (*models.ProjectSettings, error) {
	settings := &models.ProjectSettings{}
	err := db.ormer.QueryTable(constants.TableNameMap[constants.ProjectSettingsTable]).
		Filter("project_id", projectID).
		One(settings)
	if err == orm.ErrNoRows {
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

	existing := &models.ProjectSettings{ProjectID: settings.ProjectID}
	err := db.ormer.Read(existing, "ProjectID")

	if err == orm.ErrNoRows {
		settings.ID = 0
		if _, err := db.ormer.Insert(settings); err != nil {
			return fmt.Errorf("failed to insert project settings project_id[%s]: %s", settings.ProjectID, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to lookup project settings project_id[%s]: %s", settings.ProjectID, err)
	}

	// Record exists, update it
	settings.ID = existing.ID
	if _, err := db.ormer.Update(settings); err != nil {
		return fmt.Errorf("failed to update project settings project_id[%s]: %s", settings.ProjectID, err)
	}
	return nil
}
