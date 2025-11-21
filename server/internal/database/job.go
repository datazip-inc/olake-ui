package database

import (
	"fmt"

	"github.com/beego/beego/v2/client/orm"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// decryptJobConfig decrypts Config fields in related Source and Destination
func (db *Database) decryptJobConfig(job *models.Job) error {
	// Decrypt Source Config if loaded
	// TODO: verify why source_id and dest_id coming nil, it must not nil
	if job.SourceID != nil {
		decryptedConfig, err := utils.Decrypt(job.SourceID.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt source config job_id[%d] source_id[%d]: %s", job.ID, job.SourceID.ID, err)
		}
		job.SourceID.Config = decryptedConfig
	}

	// Decrypt Destination Config if loaded
	if job.DestID != nil {
		decryptedConfig, err := utils.Decrypt(job.DestID.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt destination config job_id[%d] dest_id[%d]: %s", job.ID, job.DestID.ID, err)
		}
		job.DestID.Config = decryptedConfig
	}

	return nil
}

// decryptJobSliceConfig decrypts related entities for a slice of jobs
func (db *Database) decryptJobSliceConfig(jobs []*models.Job) error {
	for _, job := range jobs {
		if err := db.decryptJobConfig(job); err != nil {
			return fmt.Errorf("failed to decrypt job config job_id[%d]: %s", job.ID, err)
		}
	}
	return nil
}

// Create a new job
func (db *Database) CreateJob(job *models.Job) error {
	_, err := db.ormer.Insert(job)
	return err
}

// GetAll retrieves all jobs
func (db *Database) ListJobs() ([]*models.Job, error) {
	var jobs []*models.Job
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).RelatedSel().OrderBy(constants.OrderByUpdatedAtDesc).All(&jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %s", err)
	}

	// Decrypt related Source and Destination configs
	if err := db.decryptJobSliceConfig(jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetAllJobsByProjectID retrieves all jobs belonging to a specific project,
// including related Source and Destination, sorted by latest update time.
func (db *Database) ListJobsByProjectID(projectID string) ([]*models.Job, error) {
	var jobs []*models.Job

	// Directly query jobs filtered by project_id — since each job already stores project_id
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).
		Filter("project_id", projectID).
		RelatedSel().
		OrderBy(constants.OrderByUpdatedAtDesc).
		All(&jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs project_id[%s]: %s", projectID, err)
	}

	// If project has no jobs, return empty slice (not nil)
	if len(jobs) == 0 {
		return []*models.Job{}, nil
	}

	// Decrypt related Source and Destination configs
	if err := db.decryptJobSliceConfig(jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetByID retrieves a job by ID
func (db *Database) GetJobByID(id int, decrypt bool) (*models.Job, error) {
	job := &models.Job{ID: id}
	err := db.ormer.Read(job)
	if err != nil {
		return nil, fmt.Errorf("failed to get job id[%d]: %s", id, err)
	}

	// Load related entities (Source, Destination, etc.)
	_, err = db.ormer.LoadRelated(job, "SourceID")
	if err != nil {
		return nil, fmt.Errorf("failed to load source entities job_id[%d]: %s", id, err)
	}

	_, err = db.ormer.LoadRelated(job, "DestID")
	if err != nil {
		return nil, fmt.Errorf("failed to load destination entities job_id[%d]: %s", id, err)
	}

	// Decrypt related Source and Destination configs
	if decrypt {
		if err := db.decryptJobConfig(job); err != nil {
			return nil, err
		}
	}

	return job, nil
}

func (db *Database) GetJobsBySourceID(sourceIDs []int) ([]*models.Job, error) {
	var jobs []*models.Job
	if len(sourceIDs) == 0 {
		return jobs, nil
	}
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).Filter("source_id__in", sourceIDs).RelatedSel().All(&jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (db *Database) GetJobsByDestinationID(destIDs []int) ([]*models.Job, error) {
	var jobs []*models.Job
	if len(destIDs) == 0 {
		return jobs, nil
	}
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).Filter("dest_id__in", destIDs).RelatedSel().All(&jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// Update a job
func (db *Database) UpdateJob(job *models.Job) error {
	_, err := db.ormer.Update(job)
	return err
}

// BulkDeactivate deactivates multiple jobs by their IDs in a single query
func (db *Database) DeactivateJobs(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).
		Filter("id__in", ids).
		Update(orm.Params{
			"active": false,
		})
	return err
}

// Delete a job
func (db *Database) DeleteJob(id int) error {
	_, err := db.ormer.Delete(&models.Job{ID: id})
	return err
}

// IsNameUniqueInProject checks if a name is unique within a project for a given table.
func (db *Database) IsNameUniqueInProject(projectID, name string, tableType constants.TableType) (bool, error) {
	tableName, ok := constants.TableNameMap[tableType]
	if !ok {
		return false, fmt.Errorf("invalid table type: %v", tableType)
	}

	count, err := db.ormer.QueryTable(tableName).
		Filter("name", name).
		Filter("project_id", projectID).
		Count()
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness project_id[%s] name[%s] table[%s]: %s", projectID, name, tableName, err)
	}
	return count == 0, nil
}

// IsJobNameUniqueInProject checks if a job name is unique within a project.
func (db *Database) IsJobNameUniqueInProject(projectID, jobName string) (bool, error) {
	return db.IsNameUniqueInProject(projectID, jobName, constants.JobTable)
}

// IsSourceNameUniqueInProject checks if a source name is unique within a project.
func (db *Database) IsSourceNameUniqueInProject(projectID, name string) (bool, error) {
	return db.IsNameUniqueInProject(projectID, name, constants.SourceTable)
}

// IsDestinationNameUniqueInProject checks if a destination name is unique within a project.
func (db *Database) IsDestinationNameUniqueInProject(projectID, name string) (bool, error) {
	return db.IsNameUniqueInProject(projectID, name, constants.DestinationTable)
}
