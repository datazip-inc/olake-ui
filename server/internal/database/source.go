package database

import (
	"fmt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/utils"
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
	_, err = db.ormer.Insert(source)
	return err
}

func (db *Database) ListSources() ([]*models.Source, error) {
	var sources []*models.Source
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.SourceTable]).RelatedSel().OrderBy(constants.OrderByUpdatedAtDesc).All(&sources)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %s", err)
	}

	// Decrypt config after reading
	if err := db.decryptSourceSliceConfigs(sources); err != nil {
		return nil, err
	}

	return sources, nil
}

func (db *Database) GetSourceByID(id int) (*models.Source, error) {
	source := &models.Source{ID: id}
	err := db.ormer.Read(source)
	if err != nil {
		return nil, fmt.Errorf("failed to get source id[%d]: %s", id, err)
	}

	// Decrypt config after reading
	dConfig, err := utils.Decrypt(source.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt source config id[%d]: %s", source.ID, err)
	}
	source.Config = dConfig
	return source, nil
}

func (db *Database) UpdateSource(source *models.Source) error {
	// Encrypt config before saving
	eConfig, err := utils.Encrypt(source.Config)
	if err != nil {
		return fmt.Errorf("failed to encrypt source config id[%d]: %s", source.ID, err)
	}
	source.Config = eConfig
	_, err = db.ormer.Update(source)
	return err
}

func (db *Database) DeleteSource(id int) error {
	source := &models.Source{ID: id}
	_, err := db.ormer.Delete(source)
	return err
}

// GetSourcesByNameAndType retrieves sources by name, type, and project ID
// func (db *Database) GetSourcesByNameAndType(name, sourceType, projectID string) ([]*models.Source, error) {
// 	var sources []*models.Source
// 	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.SourceTable]).
// 		Filter("name", name).
// 		Filter("type", sourceType).
// 		Filter("project_id", projectID).
// 		All(&sources)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get sources project_id[%s], name[%s], type[%s]: %s", projectID, name, sourceType, err)
// 	}

// 	// Decrypt config after reading
// 	if err := db.decryptSourceSliceConfigs(sources); err != nil {
// 		return nil, err
// 	}

// 	return sources, nil
// }

// // GetByNameAndType retrieves sources by name, type, and project ID
// func (db *Database) GetByNameAndType(name, sourceType, projectIDStr string) ([]*models.Source, error) {
// 	var sources []*models.Source
// 	_, err := db.ormer.QueryTable(r.TableName).
// 		Filter("name", name).
// 		Filter("type", sourceType).
// 		Filter("project_id", projectIDStr).
// 		All(&sources)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get source by name: %s and type: %s and project_id: %s: %s", name, sourceType, projectIDStr, err)
// 	}

// 	// Decrypt config after reading
// 	if err := r.decryptSourceSliceConfigs(sources); err != nil {
// 		return nil, fmt.Errorf("failed to decrypt source config: %s", err)
// 	}

// 	return sources, nil
// }
