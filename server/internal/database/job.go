package database

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/utils"
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

// decryptJobConfig decrypts Config fields in related Source and Destination
func (r *JobORM) decryptJobConfig(job *models.Job) error {
	// Decrypt Source Config if loaded
	// TODO: verify why source_id and dest_id coming nil, it must not nil
	if job.SourceID != nil {
		decryptedConfig, err := utils.DecryptConfig(job.SourceID.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt source config: %s", err)
		}
		job.SourceID.Config = decryptedConfig
	}

	// Decrypt Destination Config if loaded
	if job.DestID != nil {
		decryptedConfig, err := utils.DecryptConfig(job.DestID.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt destination config: %s", err)
		}
		job.DestID.Config = decryptedConfig
	}

	return nil
}

// decryptJobSliceConfig decrypts related entities for a slice of jobs
func (r *JobORM) decryptJobSliceConfig(jobs []*models.Job) error {
	for _, job := range jobs {
		if err := r.decryptJobConfig(job); err != nil {
			return fmt.Errorf("failed to decrypt job config: %s", err)
		}
	}
	return nil
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
	if err != nil {
		return nil, fmt.Errorf("failed to get all jobs: %s", err)
	}

	// Decrypt related Source and Destination configs
	if err := r.decryptJobSliceConfig(jobs); err != nil {
		return nil, fmt.Errorf("failed to decrypt job config: %s", err)
	}

	return jobs, nil
}

// GetAllByProjectID retrieves all jobs for a specific project
func (r *JobORM) GetAllByProjectID(projectID string) ([]*models.Job, error) {
	var jobs []*models.Job

	// Query sources in the project
	sourceTable := constants.TableNameMap[constants.SourceTable]
	sources := []int{}
	_, err := r.ormer.Raw(fmt.Sprintf(`SELECT id FROM %q WHERE project_id = ?`, sourceTable), projectID).QueryRows(&sources)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources for project ID %s: %s", projectID, err)
	}

	// Query destinations in the project
	destTable := constants.TableNameMap[constants.DestinationTable]
	destinations := []int{}
	_, err = r.ormer.Raw(fmt.Sprintf(`SELECT id FROM %q WHERE project_id = ?`, destTable), projectID).QueryRows(&destinations)
	if err != nil {
		return nil, fmt.Errorf("failed to get destinations for project ID %s: %s", projectID, err)
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
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs with related data for project ID %s: %s", projectID, err)
	}

	// Decrypt related Source and Destination configs
	if err := r.decryptJobSliceConfig(jobs); err != nil {
		return nil, fmt.Errorf("failed to decrypt job config: %s", err)
	}

	return jobs, nil
}

// GetByID retrieves a job by ID
func (r *JobORM) GetByID(id int, decrypt bool) (*models.Job, error) {
	job := &models.Job{ID: id}
	err := r.ormer.Read(job)
	if err != nil {
		return nil, fmt.Errorf("failed to get job by ID: %s", err)
	}

	// Load related entities (Source, Destination, etc.)
	_, err = r.ormer.LoadRelated(job, "SourceID")
	if err != nil {
		return nil, fmt.Errorf("failed to get job by ID: %s", err)
	}
	_, err = r.ormer.LoadRelated(job, "DestID")
	if err != nil {
		return nil, fmt.Errorf("failed to get job by ID: %s", err)
	}

	// Decrypt related Source and Destination configs
	if decrypt {
		if err := r.decryptJobConfig(job); err != nil {
			return nil, fmt.Errorf("failed to decrypt job config: %s", err)
		}
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
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source ID: %s", err)
	}

	// Decrypt related Source and Destination configs
	if err := r.decryptJobSliceConfig(jobs); err != nil {
		return nil, fmt.Errorf("failed to decrypt job config: %s", err)
	}

	return jobs, nil
}

// GetByDestinationID retrieves all jobs associated with a destination ID
func (r *JobORM) GetByDestinationID(destID int) ([]*models.Job, error) {
	var jobs []*models.Job
	dest := &models.Destination{ID: destID}

	_, err := r.ormer.QueryTable(r.TableName).
		Filter("dest_id", dest).
		RelatedSel().
		All(&jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by destination ID: %s", err)
	}

	// Decrypt related Source and Destination configs
	if err := r.decryptJobSliceConfig(jobs); err != nil {
		return nil, fmt.Errorf("failed to decrypt job config: %s", err)
	}

	return jobs, nil
}
