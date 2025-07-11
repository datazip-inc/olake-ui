package service

import (
	"olake-ui/olake-workers/k8s/database"
	"olake-ui/olake-workers/k8s/logger"
)

// JobDataService defines the interface for job data operations
type JobDataService interface {
	GetJobData(jobID int) (*database.JobData, error)
	UpdateJobState(jobID int, state string, active bool) error
	Close() error
}

// PostgresJobService implements JobDataService using PostgreSQL
type PostgresJobService struct {
	db *database.DB
}

// NewPostgresJobService creates a new PostgreSQL job service
func NewPostgresJobService() (*PostgresJobService, error) {
	db, err := database.NewDB()
	if err != nil {
		logger.Errorf("Failed to create database connection: %v", err)
		return nil, err
	}

	logger.Info("Created PostgreSQL job service")
	return &PostgresJobService{db: db}, nil
}

// GetJobData retrieves job configuration from database
func (s *PostgresJobService) GetJobData(jobID int) (*database.JobData, error) {
	logger.Debugf("Getting job data for jobID: %d", jobID)
	return s.db.GetJobData(jobID)
}

// UpdateJobState updates the job state and active status in the database
func (s *PostgresJobService) UpdateJobState(jobID int, state string, active bool) error {
	logger.Debugf("Updating job state for jobID: %d", jobID)
	return s.db.UpdateJobState(jobID, state, active)
}

// Close closes the database connection
func (s *PostgresJobService) Close() error {
	logger.Debug("Closing PostgreSQL job service")
	return s.db.Close()
}
