package database

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
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
	var sources []*models.Source
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.SourceTable]).
		Filter("id", id).
		RelatedSel().
		All(&sources)
	if err != nil {
		return nil, fmt.Errorf("failed to get source id[%d]: %s", id, err)
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("source not found id[%d]", id)
	}

	source := sources[0]

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

// IsSourceNameUniqueInProject checks if a source name is unique within a project.
func (db *Database) IsSourceNameUniqueInProject(ctx context.Context, projectID, name string) (bool, error) {
	return db.IsNameUniqueInProject(ctx, projectID, name, constants.SourceTable)
}
