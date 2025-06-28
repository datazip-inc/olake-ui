package database

import (
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/crypto"
	"github.com/datazip/olake-frontend/server/internal/models"
)

// DestinationORM handles database operations for destinations
type DestinationORM struct {
	ormer     orm.Ormer
	TableName string
}

func NewDestinationORM() *DestinationORM {
	return &DestinationORM{
		ormer:     orm.NewOrm(),
		TableName: constants.TableNameMap[constants.DestinationTable],
	}
}

// encryptDestinationConfig encrypts the config field before saving
func (r *DestinationORM) encryptDestinationConfig(destination *models.Destination) error {
	if destination.Config != "" {
		encryptedConfig, err := crypto.EncryptJSONString(destination.Config)
		if err != nil {
			return err
		}
		destination.Config = encryptedConfig
	}
	return nil
}

// decryptDestinationConfig decrypts the config field after reading
func (r *DestinationORM) decryptDestinationConfig(destination *models.Destination) error {
	if destination.Config != "" {
		decryptedConfig, err := crypto.DecryptJSONString(destination.Config)
		if err != nil {
			return err
		}
		destination.Config = decryptedConfig
	}
	return nil
}

// decryptDestinationSliceConfigs decrypts config fields for a slice of destinations
func (r *DestinationORM) decryptDestinationSliceConfigs(destinations []*models.Destination) error {
	for _, dest := range destinations {
		if err := r.decryptDestinationConfig(dest); err != nil {
			return err
		}
	}
	return nil
}

func (r *DestinationORM) Create(destination *models.Destination) error {
	// Encrypt config before saving
	if err := r.encryptDestinationConfig(destination); err != nil {
		return err
	}

	_, err := r.ormer.Insert(destination)
	return err
}

func (r *DestinationORM) GetAll() ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&destinations)
	if err != nil {
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, err
	}

	return destinations, nil
}

func (r *DestinationORM) GetAllByProjectID(projectID string) ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).Filter("project_id", projectID).RelatedSel().All(&destinations)
	if err != nil {
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, err
	}

	return destinations, nil
}

func (r *DestinationORM) GetByID(id int) (*models.Destination, error) {
	destination := &models.Destination{ID: id}
	err := r.ormer.Read(destination)
	if err != nil {
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptDestinationConfig(destination); err != nil {
		return nil, err
	}

	return destination, nil
}

func (r *DestinationORM) Update(destination *models.Destination) error {
	destination.UpdatedAt = time.Now()

	// Encrypt config before saving
	if err := r.encryptDestinationConfig(destination); err != nil {
		return err
	}

	_, err := r.ormer.Update(destination)
	return err
}

func (r *DestinationORM) Delete(id int) error {
	destination := &models.Destination{ID: id}
	_, err := r.ormer.Delete(destination)
	return err
}

// GetByNameAndType retrieves destinations by name, type, and project ID
func (r *DestinationORM) GetByNameAndType(name, destType, projectIDStr string) ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).
		Filter("name", name).
		Filter("dest_type", destType).
		Filter("project_id", projectIDStr).
		All(&destinations)
	if err != nil {
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, err
	}

	return destinations, nil
}
