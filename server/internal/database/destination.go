package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
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

func (r *DestinationORM) Create(destination *models.Destination) error {
	_, err := r.ormer.Insert(destination)
	if err != nil {
		return err
	}

	// Track destination creation event
	properties := map[string]interface{}{
		"destination_id":   destination.ID,
		"destination_name": destination.Name,
		"destination_type": destination.DestType,
		"version":          destination.Version,
	}

	properties["catalog"] = "none"
	var config map[string]interface{}
	// parse config to get catalog_type
	if err := json.Unmarshal([]byte(destination.Config), &config); err == nil {
		if writer, exists := config["writer"].(map[string]interface{}); exists {
			if catalogType, exists := writer["catalog_type"]; exists {
				properties["catalog"] = catalogType
			}
		}
	}

	if destination.CreatedBy != nil {
		userORM := NewUserORM()
		if fullUser, err := userORM.GetByID(destination.CreatedBy.ID); err == nil {
			properties["created_by"] = fullUser.Username
		}
	}
	if !destination.CreatedAt.IsZero() {
		properties["created_at"] = destination.CreatedAt.Format(time.RFC3339)
	}

	if err := telemetry.TrackEvent(context.TODO(), constants.EventDestinationCreated, properties); err != nil {
		orm.DebugLog.Println("Failed to track destination creation event:", err)
	}

	return nil
}

func (r *DestinationORM) GetAll() ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&destinations)
	return destinations, err
}

func (r *DestinationORM) GetAllByProjectID(projectID string) ([]*models.Destination, error) {
	var destinations []*models.Destination
	_, err := r.ormer.QueryTable(r.TableName).Filter("project_id", projectID).RelatedSel().All(&destinations)
	return destinations, err
}

func (r *DestinationORM) GetByID(id int) (*models.Destination, error) {
	destination := &models.Destination{ID: id}
	err := r.ormer.Read(destination)
	return destination, err
}

func (r *DestinationORM) Update(destination *models.Destination) error {
	destination.UpdatedAt = time.Now()
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
	return destinations, err
}
