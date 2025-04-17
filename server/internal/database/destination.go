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
	_, err := r.ormer.QueryTable(r.TableName).All(&destinations)
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
