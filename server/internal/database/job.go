package database

import (
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/models"
)

// JobORM handles database operations for jobs
type JobORM struct {
	ormer     orm.Ormer
	TableName string
}

// NewJobORM creates a new instance of JobORM
func NewJobORM() *JobORM {
	return &JobORM{
		ormer:     orm.NewOrm(),
		TableName: constants.TableNameMap[constants.JobTable],
	}
}

// Create a new job
func (r *JobORM) Create(job *models.Job) error {
	_, err := r.ormer.Insert(job)
	return err
}

// GetAll retrieves all jobs
func (r *JobORM) GetAll() ([]*models.Job, error) {
	var jobs []*models.Job
	_, err := r.ormer.QueryTable(r.TableName).All(&jobs)
	return jobs, err
}

// GetByID retrieves a job by ID
func (r *JobORM) GetByID(id int) (*models.Job, error) {
	job := &models.Job{ID: id}
	err := r.ormer.Read(job)
	return job, err
}

// Update a job
func (r *JobORM) Update(job *models.Job) error {
	job.UpdatedAt = time.Now()
	_, err := r.ormer.Update(job)
	return err
}

// Delete a job
func (r *JobORM) Delete(id int) error {
	job := &models.Job{ID: id}
	_, err := r.ormer.Delete(job)
	return err
}
