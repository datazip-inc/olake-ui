package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"olake-k8s-worker/logger"

	_ "github.com/lib/pq"
)

type DB struct {
	conn       *sql.DB
	tableNames map[string]string
}

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

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	dbURL := getDBURL()

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize table names based on environment
	tableNames := getTableNames()

	logger.Info("Successfully connected to PostgreSQL database")

	return &DB{
		conn:       conn,
		tableNames: tableNames,
	}, nil
}

// GetJobData retrieves job configuration from database
func (db *DB) GetJobData(jobID int) (*JobData, error) {
	query := fmt.Sprintf(`
		SELECT 
			j.id, j.name, j.streams_config, j.state, j.project_id, j.active,
			s.type as source_type, s.version as source_version, s.config as source_config,
			d.type as dest_type, d.version as dest_version, d.config as dest_config
		FROM %s j
		JOIN %s s ON j.source_id = s.id
		JOIN %s d ON j.dest_id = d.id
		WHERE j.id = $1
	`, db.tableNames["job"], db.tableNames["source"], db.tableNames["destination"])

	row := db.conn.QueryRow(query, jobID)

	var jobData JobData
	err := row.Scan(
		&jobData.ID,
		&jobData.Name,
		&jobData.StreamsConfig,
		&jobData.State,
		&jobData.ProjectID,
		&jobData.Active,
		&jobData.SourceType,
		&jobData.SourceVersion,
		&jobData.SourceConfig,
		&jobData.DestType,
		&jobData.DestVersion,
		&jobData.DestConfig,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job with ID %d not found", jobID)
		}
		return nil, fmt.Errorf("failed to query job data: %w", err)
	}

	return &jobData, nil
}

// UpdateJobState updates the job state in database
func (db *DB) UpdateJobState(jobID int, state map[string]interface{}) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s 
		SET state = $1, updated_at = $2 
		WHERE id = $3
	`, db.tableNames["job"])

	_, err = db.conn.Exec(query, string(stateJSON), time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to update job state: %w", err)
	}

	logger.Infof("Updated job state for job ID: %d", jobID)
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Helper functions

func getDBURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to individual components
		host := getEnvWithDefault("DB_HOST", "localhost")
		port := getEnvWithDefault("DB_PORT", "5432")
		user := getEnvWithDefault("DB_USER", "olake")
		password := getEnvWithDefault("DB_PASSWORD", "password")
		dbname := getEnvWithDefault("DB_NAME", "olake")
		sslmode := getEnvWithDefault("DB_SSLMODE", "disable")

		dbURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}
	return dbURL
}

func getTableNames() map[string]string {
	// Get environment mode (development, production, etc.)
	runMode := getEnvWithDefault("RUN_MODE", "development")

	// Table name patterns matching server implementation
	tableNames := map[string]string{
		"user":        fmt.Sprintf("olake-%s-user", runMode),
		"source":      fmt.Sprintf("olake-%s-source", runMode),
		"destination": fmt.Sprintf("olake-%s-destination", runMode),
		"job":         fmt.Sprintf("olake-%s-job", runMode),
		"catalog":     fmt.Sprintf("olake-%s-catalog", runMode),
		"session":     "session",
	}

	return tableNames
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
