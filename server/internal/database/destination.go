package database

import (
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/models"
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
	return err
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
