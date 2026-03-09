package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// jobListColumns is the minimal Job projection needed by job list responses.
// It intentionally omits heavy fields like streams_config and state.
var jobListColumns = []string{
	"id",
	"name",
	"frequency",
	"active",
	"created_at",
	"updated_at",
	"source_id",
	"dest_id",
	"created_by_id",
	"updated_by_id",
	"project_id",
	"advanced_settings",
}

// decryptJobConfig decrypts Config fields in related Source and Destination
func (db *Database) decryptJobConfig(job *models.Job) error {
	// Decrypt Source Config if loaded
	if job.Source != nil {
		decryptedConfig, err := utils.Decrypt(job.Source.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt source config job_id[%d] source_id[%d]: %s", job.ID, job.Source.ID, err)
		}
		job.Source.Config = decryptedConfig
	}

	// Decrypt Destination Config if loaded
	if job.Destination != nil {
		decryptedConfig, err := utils.Decrypt(job.Destination.Config)
		if err != nil {
			return fmt.Errorf("failed to decrypt destination config job_id[%d] dest_id[%d]: %s", job.ID, job.Destination.ID, err)
		}
		job.Destination.Config = decryptedConfig
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
	return db.conn.Create(job).Error
}

// GetAll retrieves all jobs
func (db *Database) ListJobs() ([]*models.Job, error) {
	var jobs []*models.Job
	err := db.conn.
		Preload("Source").
		Preload("Destination").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Order("updated_at DESC").
		Find(&jobs).Error
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

	err := db.conn.
		Model(&models.Job{}).
		Select(jobListColumns).
		Where("project_id = ?", projectID).
		Preload("Source").
		Preload("Destination").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Order("updated_at DESC").
		Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs project_id[%s]: %s", projectID, err)
	}

	// If project has no jobs, return empty slice (not nil)
	if len(jobs) == 0 {
		return []*models.Job{}, nil
	}

	return jobs, nil
}

// GetByID retrieves a job by ID
func (db *Database) GetJobByID(id int, decrypt bool) (*models.Job, error) {
	job := &models.Job{}
	err := db.conn.
		Where("id = ?", id).
		Preload("Source").
		Preload("Destination").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		First(job).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get job id[%d]: %s", id, err)
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
	err := db.conn.
		Where("source_id IN ?", sourceIDs).
		Preload("Source").
		Preload("Destination").
		Find(&jobs).Error
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
	err := db.conn.
		Where("dest_id IN ?", destIDs).
		Preload("Source").
		Preload("Destination").
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// UpdateJob updates a job with the given params.
func (db *Database) UpdateJob(jobID int, params map[string]any) error {
	return db.conn.Model(&models.Job{}).
		Where("id = ?", jobID).
		Updates(params).Error
}

// BulkDeactivate deactivates multiple jobs by their IDs in a single query
func (db *Database) DeactivateJobs(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	return db.conn.Model(&models.Job{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"active": false}).Error
}

// Delete a job
func (db *Database) DeleteJob(id int) error {
	result := db.conn.Delete(&models.Job{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// IsNameUniqueInProject checks if a name is unique within a project for a given table.
func (db *Database) IsNameUniqueInProject(ctx context.Context, projectID, name string, tableType constants.TableType) (bool, error) {
	tableName, ok := constants.TableNameMap[tableType]
	if !ok {
		return false, fmt.Errorf("invalid table type: %v", tableType)
	}

	countQuery := db.conn.Table(tableName)
	if ctx != nil {
		countQuery = countQuery.WithContext(ctx)
	}

	var count int64
	err := countQuery.
		Where("name = ?", name).
		Where("project_id = ?", projectID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness project_id[%s] name[%s] table[%s]: %s", projectID, name, tableName, err)
	}
	return count == 0, nil
}

// IsJobNameUniqueInProject checks if a job name is unique within a project.
func (db *Database) IsJobNameUniqueInProject(ctx context.Context, projectID, jobName string) (bool, error) {
	return db.IsNameUniqueInProject(ctx, projectID, jobName, constants.JobTable)
}
