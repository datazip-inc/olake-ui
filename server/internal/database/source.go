package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"golang.org/x/mod/semver"
)

// decryptSourceSliceConfigs decrypts config fields for a slice of sources
func (db *Database) decryptSourceSliceConfigs(sources []*models.Source) error {
	for _, source := range sources {
		dConfig, err := utils.Decrypt(source.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt source config id[%d]: %s", source.ID, err)
		}
		source.Config = dConfig
	}
	return nil
}

func (db *Database) CreateSource(source *models.Source) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(source.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt source config id[%d]: %s", source.ID, err)
	}
	source.Config = eConfig
	return db.conn.Create(source).Error
}

func (db *Database) ListSources() ([]*models.Source, error) {
	var sources []*models.Source
	err := db.conn.
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Order("updated_at DESC").
		Find(&sources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %s", err)
	}

	// Decrypt config after reading
	if err := db.decryptSourceSliceConfigs(sources); err != nil {
		return nil, err
	}

	return sources, nil
}

func (db *Database) ListSourcesByProjectID(projectID string) ([]*models.Source, error) {
	var sources []*models.Source
	err := db.conn.
		Where("project_id = ?", projectID).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Order("updated_at DESC").
		Find(&sources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list sources project_id[%s]: %s", projectID, err)
	}

	// Decrypt config after reading
	if err := db.decryptSourceSliceConfigs(sources); err != nil {
		return nil, err
	}

	return sources, nil
}

func (db *Database) GetSourceByID(id int) (*models.Source, error) {
	var source models.Source
	err := db.conn.
		Where("id = ?", id).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&source).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: source not found id[%d]", constants.ErrSourceNotFound, id)
		}
		return nil, fmt.Errorf("failed to get source id[%d]: %s", id, err)
	}

	// Decrypt config after reading
	dConfig, err := utils.Decrypt(source.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt source config id[%d]: %s", source.ID, err)
	}
	source.Config = dConfig
	return &source, nil
}

func (db *Database) UpdateSource(source *models.Source) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(source.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt source config id[%d]: %s", source.ID, err)
	}
	source.Config = eConfig
	return db.conn.Updates(source).Error
}

func (db *Database) DeleteSource(id int) error {
	result := db.conn.Delete(&models.Source{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return constants.ErrSourceNotFound
	}
	return nil
}

// IsSourceNameUniqueInProject checks if a source name is unique within a project.
func (db *Database) IsSourceNameUniqueInProject(ctx context.Context, projectID, name string) (bool, error) {
	return db.IsNameUniqueInProject(ctx, projectID, name, constants.SourceTable)
}

// GetMinimumSourceVersion returns the minimum (oldest) semantic version from all sources.
// Uses semver comparison to find the true minimum version.
// Returns empty string if no sources exist.
func (db *Database) GetMinimumSourceVersion() (string, error) {
	var versions []string
	err := db.conn.
		Model(&models.Source{}).
		Distinct().
		Pluck("version", &versions).Error
	if err != nil {
		return "", fmt.Errorf("failed to get source versions: %s", err)
	}

	if len(versions) == 0 {
		return "", nil
	}

	// Find minimum version using semver comparison
	minVersion := ""
	for _, v := range versions {
		version := v
		if version == "" {
			continue
		}
		if minVersion == "" || semver.Compare(version, minVersion) < 0 {
			minVersion = version
		}
	}

	return minVersion, nil
}
