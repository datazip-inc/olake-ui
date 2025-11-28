package database

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// encryptProjectSettings encrypts secrets before persistence.
func (db *Database) encryptProjectSettings(settings *models.ProjectSettings) error {
	if settings == nil {
		return fmt.Errorf("project settings payload is nil")
	}

	if strings.TrimSpace(settings.WebhookAlertURL) == "" {
		return nil
	}

	encryptedURL, err := utils.Encrypt(settings.WebhookAlertURL)
	if err != nil {
		return fmt.Errorf("failed to encrypt webhook url project_id[%s]: %s", settings.ProjectID, err)
	}

	settings.WebhookAlertURL = encryptedURL
	return nil
}

// decryptProjectSettings decrypts sensitive fields after fetch.
func (db *Database) decryptProjectSettings(settings *models.ProjectSettings) error {
	if settings == nil || strings.TrimSpace(settings.WebhookAlertURL) == "" {
		return nil
	}

	decryptedURL, err := utils.Decrypt(settings.WebhookAlertURL)
	if err != nil {
		return fmt.Errorf("failed to decrypt webhook url project_id[%s]: %s", settings.ProjectID, err)
	}

	settings.WebhookAlertURL = decryptedURL
	return nil
}

// GetProjectSettingsByProjectID fetches the settings row for a project ID.
func (db *Database) GetProjectSettingsByProjectID(projectID string) (*models.ProjectSettings, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

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

	if err := db.decryptProjectSettings(settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateProjectSettingsModel upserts (inserts or updates) project settings.
func (db *Database) UpsertProjectSettingsModel(settings *models.ProjectSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}

	// 1. Encrypt the data before saving
	if err := db.encryptProjectSettings(settings); err != nil {
		return err
	}

	existing := &models.ProjectSettings{ProjectID: settings.ProjectID}
	err := db.ormer.Read(existing, "ProjectID")

	switch err {
	case orm.ErrNoRows:
		settings.ID = 0
		if _, err := db.ormer.Insert(settings); err != nil {
			return fmt.Errorf("failed to insert project settings project_id[%s]: %s", settings.ProjectID, err)
		}
	case nil:
		if _, err := db.ormer.Update(settings); err != nil {
			return fmt.Errorf("failed to update project settings project_id[%s]: %s", settings.ProjectID, err)
		}
	default:
		return fmt.Errorf("failed to lookup project settings project_id[%s]: %s", settings.ProjectID, err)
	}

	return nil
}
