package database

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/utils"
)

// SourceORM handles database operations for sources
type SourceORM struct {
	ormer     orm.Ormer
	TableName string
}

func NewSourceORM() *SourceORM {
	return &SourceORM{
		ormer:     orm.NewOrm(),
		TableName: constants.TableNameMap[constants.SourceTable],
	}
}

// decryptSourceSliceConfigs decrypts config fields for a slice of sources
func (r *SourceORM) decryptSourceSliceConfigs(sources []*models.Source) error {
	for _, source := range sources {
		dConfig, err := utils.Decrypt(source.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt source config: %s", err)
		}
		source.Config = dConfig
	}
	return nil
}

func (r *SourceORM) Create(source *models.Source) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(source.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt source config: %s", err)
	}
	source.Config = eConfig
	_, err = r.ormer.Insert(source)
	return err
}

func (r *SourceORM) GetAll() ([]*models.Source, error) {
	var sources []*models.Source
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&sources)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sources: %s", err)
	}

	// Decrypt config after reading
	if err := r.decryptSourceSliceConfigs(sources); err != nil {
		return nil, fmt.Errorf("failed to decrypt source config: %s", err)
	}

	return sources, nil
}

func (r *SourceORM) GetByID(id int) (*models.Source, error) {
	source := &models.Source{ID: id}
	err := r.ormer.Read(source)
	if err != nil {
		return nil, fmt.Errorf("failed to get source by id[%d]: %s", id, err)
	}

	// Decrypt config after reading
	dConfig, err := utils.Decrypt(source.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt source config by id[%d]: %s", source.ID, err)
	}
	source.Config = dConfig
	return source, nil
}

func (r *SourceORM) Update(source *models.Source) error {
	// TODO: remove all code managed db timestamps
	source.UpdatedAt = time.Now()
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(source.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt source config: %s", err)
	}
	source.Config = eConfig
	_, err = r.ormer.Update(source)
	return err
}

func (r *SourceORM) Delete(id int) error {
	source := &models.Source{ID: id}
	_, err := r.ormer.Delete(source)
	return err
}

// GetByNameAndType retrieves sources by name, type, and project ID
func (r *SourceORM) GetByNameAndType(name, sourceType, projectIDStr string) ([]*models.Source, error) {
	var sources []*models.Source
	_, err := r.ormer.QueryTable(r.TableName).
		Filter("name", name).
		Filter("type", sourceType).
		Filter("project_id", projectIDStr).
		All(&sources)
	if err != nil {
		return nil, fmt.Errorf("failed to get source by name: %s and type: %s and project_id: %s: %s", name, sourceType, projectIDStr, err)
	}

	// Decrypt config after reading
	if err := r.decryptSourceSliceConfigs(sources); err != nil {
		return nil, fmt.Errorf("failed to decrypt source config: %s", err)
	}

	return sources, nil
}
