package database

import (
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
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

	// Query sources in the project using ORM
	var sources []models.Source
	sourceQs := r.ormer.QueryTable(constants.TableNameMap[constants.SourceTable])
	_, err := sourceQs.Filter("project_id", projectID).All(&sources)
	if err != nil {
		return nil, err
	}

	// Query destinations in the project using ORM
	var destinations []models.Destination
	destQs := r.ormer.QueryTable(constants.TableNameMap[constants.DestinationTable])
	_, err = destQs.Filter("project_id", projectID).All(&destinations)
	if err != nil {
		return nil, err
	}

	// If no sources or destinations in the project, return empty array
	if len(sources) == 0 && len(destinations) == 0 {
		return jobs, nil
	}

	// Extract IDs for filtering
	sourceIDs := make([]int, len(sources))
	for i := range sources {
		sourceIDs[i] = sources[i].ID
	}

	destIDs := make([]int, len(destinations))
	for i := range destinations {
		destIDs[i] = destinations[i].ID
	}

	// Build query for jobs
	qs := r.ormer.QueryTable(r.TableName)

	// Create OR condition for sources and destinations
	cond := orm.NewCondition()
	if len(sourceIDs) > 0 {
		cond = cond.Or("source_id__in", sourceIDs)
	}
	if len(destIDs) > 0 {
		cond = cond.Or("dest_id__in", destIDs)
	}

	// Apply condition and load related objects
	_, err = qs.SetCond(cond).RelatedSel().All(&jobs)
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
func (r *JobORM) GetBySourceIDs(sourceIDs []int) ([]*models.Job, error) {
	var jobs []*models.Job
	if len(sourceIDs) == 0 {
		return jobs, nil
	}
	_, err := r.ormer.QueryTable(r.TableName).Filter("source_id__in", sourceIDs).RelatedSel().All(&jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
func (r *JobORM) GetByDestinationIDs(destIDs []int) ([]*models.Job, error) {
	var jobs []*models.Job
	if len(destIDs) == 0 {
		return jobs, nil
	}
	_, err := r.ormer.QueryTable(r.TableName).Filter("destination_id__in", destIDs).RelatedSel().All(&jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
