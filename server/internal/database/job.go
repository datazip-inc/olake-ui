package database

import (
	"fmt"
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
	_, err := r.ormer.QueryTable(r.TableName).RelatedSel().All(&jobs)
	return jobs, err
}

// GetAllByProjectID retrieves all jobs for a specific project
func (r *JobORM) GetAllByProjectID(projectID string) ([]*models.Job, error) {
	var jobs []*models.Job

	// Query sources in the project
	sourceTable := constants.TableNameMap[constants.SourceTable]
	sources := []int{}
	_, err := r.ormer.Raw(fmt.Sprintf(`SELECT id FROM "%q" WHERE project_id = ?`, sourceTable), projectID).QueryRows(&sources)
	if err != nil {
		return nil, err
	}

	// Query destinations in the project
	destTable := constants.TableNameMap[constants.DestinationTable]
	destinations := []int{}
	_, err = r.ormer.Raw(fmt.Sprintf(`SELECT id FROM "%q" WHERE project_id = ?`, destTable), projectID).QueryRows(&destinations)
	if err != nil {
		return nil, err
	}

	// If no sources or destinations in the project, return empty array
	if len(sources) == 0 && len(destinations) == 0 {
		return jobs, nil
	}

	// Build query
	qs := r.ormer.QueryTable(r.TableName)
	// Filter by sources or destinations from the project
	if len(sources) > 0 {
		qs = qs.Filter("source_id__in", sources)
	}

	if len(destinations) > 0 {
		qs = qs.Filter("dest_id__in", destinations)
	}

	// Add RelatedSel to load the related Source and Destination objects
	_, err = qs.RelatedSel().All(&jobs)
	return jobs, err
}

// GetByID retrieves a job by ID
func (r *JobORM) GetByID(id int) (*models.Job, error) {
	job := &models.Job{ID: id}
	err := r.ormer.Read(job)
	if err != nil {
		return nil, err
	}

	// Load related entities (Source, Destination, etc.)
	_, err = r.ormer.LoadRelated(job, "SourceID")
	if err != nil {
		return nil, err
	}
	_, err = r.ormer.LoadRelated(job, "DestID")
	if err != nil {
		return nil, err
	}

	return job, nil
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

// GetBySourceID retrieves all jobs associated with a source ID
func (r *JobORM) GetBySourceID(sourceID int) ([]*models.Job, error) {
	var jobs []*models.Job
	source := &models.Source{ID: sourceID}

	_, err := r.ormer.QueryTable(r.TableName).
		Filter("source_id", source).
		RelatedSel().
		All(&jobs)

	return jobs, err
}

// GetByDestinationID retrieves all jobs associated with a destination ID
func (r *JobORM) GetByDestinationID(destID int) ([]*models.Job, error) {
	var jobs []*models.Job
	dest := &models.Destination{ID: destID}

	_, err := r.ormer.QueryTable(r.TableName).
		Filter("dest_id", dest).
		RelatedSel().
		All(&jobs)

	return jobs, err
}
