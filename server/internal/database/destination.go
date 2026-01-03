package database

import (
	"context"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
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
	_, err = db.ormer.Insert(destination)
	return err
}

func (db *Database) ListDestinations() ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.DestinationTable]).RelatedSel().OrderBy(constants.OrderByUpdatedAtDesc).All(&destinations)
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
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.DestinationTable]).Filter("project_id", projectID).RelatedSel().OrderBy(constants.OrderByUpdatedAtDesc).All(&destinations)
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
	err := db.ormer.QueryTable(constants.TableNameMap[constants.DestinationTable]).
		Filter("id", id).
		RelatedSel().
		One(&destination)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("destination not found id[%d]", id)
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
	_, err = db.ormer.Update(destination)
	return err
}

func (db *Database) DeleteDestination(id int) error {
	destination := &models.Destination{ID: id}
	// Use ORM's Delete method which will automatically handle the soft delete
	// by setting the DeletedAt field due to the ORM tags in BaseModel
	_, err := db.ormer.Delete(destination)
	return err
}

// IsDestinationNameUniqueInProject checks if a destination name is unique within a project.
func (db *Database) IsDestinationNameUniqueInProject(ctx context.Context, projectID, name string) (bool, error) {
	return db.IsNameUniqueInProject(ctx, projectID, name, constants.DestinationTable)
}
