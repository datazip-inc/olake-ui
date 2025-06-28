package database

import (
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/crypto"
	"github.com/datazip/olake-frontend/server/internal/models"
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

// encryptSourceConfig encrypts the config field before saving
func (r *SourceORM) encryptSourceConfig(source *models.Source) error {
	if source.Config != "" {
		encryptedConfig, err := crypto.EncryptJSONString(source.Config)
		if err != nil {
			return err
		}
		source.Config = encryptedConfig
	}
	return nil
}

// decryptSourceConfig decrypts the config field after reading
func (r *SourceORM) decryptSourceConfig(source *models.Source) error {
	if source.Config != "" {
		decryptedConfig, err := crypto.DecryptJSONString(source.Config)
		if err != nil {
			return err
		}
		source.Config = decryptedConfig
	}
	return nil
}

// decryptSourceSliceConfigs decrypts config fields for a slice of sources
func (r *SourceORM) decryptSourceSliceConfigs(sources []*models.Source) error {
	for _, source := range sources {
		if err := r.decryptSourceConfig(source); err != nil {
			return err
		}
	}
	return nil
}

func (r *SourceORM) Create(source *models.Source) error {
	// Encrypt config before saving
	if err := r.encryptSourceConfig(source); err != nil {
		return err
	}
	_, err := r.ormer.Insert(source)
	return err
}

func (r *SourceORM) GetAll() ([]*models.Source, error) {
	var sources []*models.Source
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&sources)
	if err != nil {
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptSourceSliceConfigs(sources); err != nil {
		return nil, err
	}

	return sources, nil
}

func (r *SourceORM) GetByID(id int) (*models.Source, error) {
	source := &models.Source{ID: id}
	err := r.ormer.Read(source)
	if err != nil {
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptSourceConfig(source); err != nil {
		return nil, err
	}

	return source, nil
}

func (r *SourceORM) Update(source *models.Source) error {
	source.UpdatedAt = time.Now()

	// Encrypt config before saving
	if err := r.encryptSourceConfig(source); err != nil {
		return err
	}

	_, err := r.ormer.Update(source)
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
		return nil, err
	}

	// Decrypt config after reading
	if err := r.decryptSourceSliceConfigs(sources); err != nil {
		return nil, err
	}

	return sources, nil
}
