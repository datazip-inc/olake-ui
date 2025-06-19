package database

import (
	"context"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
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

func (r *SourceORM) Create(source *models.Source) error {
	_, err := r.ormer.Insert(source)
	if err != nil {
		return err
	}

	// Track source creation event
	properties := map[string]interface{}{
		"source_id":   source.ID,
		"source_name": source.Name,
		"source_type": source.Type,
		"version":     source.Version,
	}
	if source.CreatedBy != nil {
		userORM := NewUserORM()
		if fullUser, err := userORM.GetByID(source.CreatedBy.ID); err == nil {
			properties["created_by"] = fullUser.Username
		}
	}
	if !source.CreatedAt.IsZero() {
		properties["created_at"] = source.CreatedAt.Format(time.RFC3339)
	}

	if err := telemetry.TrackEvent(context.TODO(), constants.EventSourceCreated, properties); err != nil {
		orm.DebugLog.Println("Failed to track source creation event:", err)
	}

	return nil
}

func (r *SourceORM) GetAll() ([]*models.Source, error) {
	var sources []*models.Source
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&sources)
	return sources, err
}

func (r *SourceORM) GetByID(id int) (*models.Source, error) {
	source := &models.Source{ID: id}
	err := r.ormer.Read(source)
	return source, err
}

func (r *SourceORM) Update(source *models.Source) error {
	source.UpdatedAt = time.Now()
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
	return sources, err
}
