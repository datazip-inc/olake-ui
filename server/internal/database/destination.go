package database

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/utils"
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

// decryptDestinationSliceConfigs decrypts config fields for a slice of destinations
func (r *DestinationORM) decryptDestinationSliceConfigs(destinations []*models.Destination) error {
	for _, dest := range destinations {
		dConfig, err := utils.Decrypt(dest.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt destination config: %s", err)
		}
		dest.Config = dConfig
	}
	return nil
}

func (r *DestinationORM) Create(destination *models.Destination) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(destination.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt destination config: %s", err)
	}
	destination.Config = eConfig
	_, err = r.ormer.Insert(destination)
	return err
}

func (r *DestinationORM) GetAll() ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&destinations)
	if err != nil {
		return nil, fmt.Errorf("failed to get all destinations: %s", err)
	}

	// Decrypt config after reading
	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, fmt.Errorf("failed to decrypt destination config: %s", err)
	}

	return destinations, nil
}

func (r *DestinationORM) GetAllByProjectID(projectID string) ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).Filter("project_id", projectID).RelatedSel().All(&destinations)
	if err != nil {
		return nil, fmt.Errorf("failed to get all destinations by project_id[%s]: %s", projectID, err)
	}

	// Decrypt config after reading
	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, fmt.Errorf("failed to decrypt destination config: %s", err)
	}

	return destinations, nil
}

func (r *DestinationORM) GetByID(id int) (*models.Destination, error) {
	destination := &models.Destination{ID: id}
	err := r.ormer.Read(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination by ID: %s", err)
	}

	// Decrypt config after reading
	dConfig, err := utils.Decrypt(destination.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt config for destination[%d]: %s", destination.ID, err)
	}
	destination.Config = dConfig
	return destination, nil
}

func (r *DestinationORM) Update(destination *models.Destination) error {
	destination.UpdatedAt = time.Now()

	// Encrypt config before saving
	eConfig, err := utils.Encrypt(destination.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt destination config: %s", err)
	}
	destination.Config = eConfig
	_, err = r.ormer.Update(destination)
	return err
}

func (r *DestinationORM) Delete(id int) error {
	destination := &models.Destination{ID: id}
	// Use ORM's Delete method which will automatically handle the soft delete
	// by setting the DeletedAt field due to the ORM tags in BaseModel
	_, err := r.ormer.Delete(destination)
	return err
}

// GetByNameAndType retrieves destinations by name, destType, and project ID
func (r *DestinationORM) GetByNameAndType(name, destType, projectID string) ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).
		Filter("name", name).
		Filter("dest_type", destType).
		Filter("project_id", projectID).
		All(&destinations)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination in project[%s] by name[%s] and type[%s]: %s", projectID, name, destType, err)
	}

	// Decrypt config after reading
	if err := r.decryptDestinationSliceConfigs(destinations); err != nil {
		return nil, fmt.Errorf("failed to decrypt destination config: %s", err)
	}

	return destinations, nil
}
