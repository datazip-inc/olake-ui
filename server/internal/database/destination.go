package database

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// decryptDestinationSliceConfigs decrypts config fields for a slice of destinations
func (db *Database) decryptDestinationSliceConfigs(destinations []*models.Destination) error {
	for _, dest := range destinations {
		dConfig, err := utils.Decrypt(dest.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt destination config id[%d]: %s", dest.ID, err)
		}
		dest.Config = dConfig
	}
	return nil
}

func (db *Database) CreateDestination(destination *models.Destination) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(destination.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt destination config id[%d]: %s", destination.ID, err)
	}
	destination.Config = eConfig
	return db.conn.Create(destination).Error
}

func (db *Database) ListDestinations() ([]*models.Destination, error) {
	var destinations []*models.Destination
	err := db.conn.
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Order("updated_at DESC").
		Find(&destinations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list destinations: %s", err)
	}

	// Decrypt config after reading
	if err := db.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, err
	}

	return destinations, nil
}

func (db *Database) ListDestinationsByProjectID(projectID string) ([]*models.Destination, error) {
	var destinations []*models.Destination
	err := db.conn.
		Where("project_id = ?", projectID).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Order("updated_at DESC").
		Find(&destinations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list destinations project_id[%s]: %s", projectID, err)
	}

	// Decrypt config after reading
	if err := db.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, err
	}

	return destinations, nil
}

func (db *Database) GetDestinationByID(id int) (*models.Destination, error) {
	var destination models.Destination
	err := db.conn.
		Where("id = ?", id).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(&destination).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: destination not found id[%d]", constants.ErrDestinationNotFound, id)
		}
		return nil, fmt.Errorf("failed to get destination id[%d]: %s", id, err)
	}

	// Decrypt config after reading
	dConfig, err := utils.Decrypt(destination.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt destination config id[%d]: %s", destination.ID, err)
	}
	destination.Config = dConfig
	return &destination, nil
}

func (db *Database) UpdateDestination(destination *models.Destination) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(destination.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt destination[%d] config: %s", destination.ID, err)
	}
	destination.Config = eConfig
	return db.conn.Updates(destination).Error
}

func (db *Database) DeleteDestination(id int) error {
	result := db.conn.Delete(&models.Destination{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return constants.ErrDestinationNotFound
	}
	return nil
}

// IsDestinationNameUniqueInProject checks if a destination name is unique within a project.
func (db *Database) IsDestinationNameUniqueInProject(ctx context.Context, projectID, name string) (bool, error) {
	return db.IsNameUniqueInProject(ctx, projectID, name, constants.DestinationTable)
}
