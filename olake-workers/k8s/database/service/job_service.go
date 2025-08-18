package service

import (
	"context"

	"olake-ui/olake-workers/k8s/database"
	"olake-ui/olake-workers/k8s/logger"
)

// JobData represents the job configuration data from database
type JobData struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	SourceType    string `json:"source_type"`
	SourceVersion string `json:"source_version"`
	SourceConfig  string `json:"source_config"`
	DestType      string `json:"dest_type"`
	DestVersion   string `json:"dest_version"`
	DestConfig    string `json:"dest_config"`
	StreamsConfig string `json:"streams_config"`
	State         string `json:"state"`
	ProjectID     string `json:"project_id"`
	Active        bool   `json:"active"`
}

// JobDataService defines the interface for job data operations
type JobDataService interface {
	GetJobData(ctx context.Context, jobID int) (*JobData, error)
	UpdateJobState(ctx context.Context, jobID int, state string, active bool) error
	HealthCheck() error
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

	return &PostgresJobService{db: db}, nil
}

// GetJobData retrieves job configuration from database
func (s *PostgresJobService) GetJobData(ctx context.Context, jobID int) (*JobData, error) {
	logger.Debugf("Getting job data for jobID: %d", jobID)
	dbJobData, err := s.db.GetJobData(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Convert from internal database jobData to service JobData
	return &JobData{
		ID:            dbJobData.ID,
		Name:          dbJobData.Name,
		SourceType:    dbJobData.SourceType,
		SourceVersion: dbJobData.SourceVersion,
		SourceConfig:  dbJobData.SourceConfig,
		DestType:      dbJobData.DestType,
		DestVersion:   dbJobData.DestVersion,
		DestConfig:    dbJobData.DestConfig,
		StreamsConfig: dbJobData.StreamsConfig,
		State:         dbJobData.State,
		ProjectID:     dbJobData.ProjectID,
		Active:        dbJobData.Active,
	}, nil
}

// UpdateJobState updates the job state and active status in the database
func (s *PostgresJobService) UpdateJobState(ctx context.Context, jobID int, state string, active bool) error {
	logger.Debugf("Updating job state for jobID: %d", jobID)
	return s.db.UpdateJobState(ctx, jobID, state, active)
}

// HealthCheck pings the database to verify connectivity
func (s *PostgresJobService) HealthCheck() error {
	return s.db.Ping()
}

// Close closes the database connection
func (s *PostgresJobService) Close() error {
	logger.Debug("Closing PostgreSQL job service")
	return s.db.Close()
}
