package database

import (
	"fmt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/utils"
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
	destination := &models.Destination{ID: id}
	err := db.ormer.Read(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination id[%d]: %s", id, err)
	}

	// Decrypt config after reading
	dConfig, err := utils.Decrypt(destination.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt destination config id[%d]: %s", destination.ID, err)
	}
	destination.Config = dConfig
	return destination, nil
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

// // GetDestinationsByNameAndType retrieves destinations by name, destType, and project ID
// func (db *Database) GetDestinationsByNameAndType(name, destType, projectID string) ([]*models.Destination, error) {
// 	var destinations []*models.Destination
// 	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.DestinationTable]).
// 		Filter("name", name).
// 		Filter("dest_type", destType).
// 		Filter("project_id", projectID).
// 		All(&destinations)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get destinations project_id[%s], name[%s], type[%s]: %s", projectID, name, destType, err)
// 	}

// 	// Decrypt config after reading
// 	if err := db.decryptDestinationSliceConfigs(destinations); err != nil {
// 		return nil, err
// 	}

// 	return destinations, nil
// }

// // GetByNameAndType retrieves destinations by name, destType, and project ID
// func (db *Database) GetDestinationsByNameAndType(name, destType, projectID string) ([]*models.Destination, error) {
// 	var destinations []*models.Destination
// 	_, err := db.ormer.QueryTable(r.TableName).
// 		Filter("name", name).
// 		Filter("dest_type", destType).
// 		Filter("project_id", projectID).
// 		All(&destinations)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get destination in project[%s] by name[%s] and type[%s]: %s", projectID, name, destType, err)
// 	}

// 	// Decrypt config after reading
// 	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
// 		return nil, fmt.Errorf("failed to decrypt destination config: %s", err)
// 	}

// 	return destinations, nil
// }
