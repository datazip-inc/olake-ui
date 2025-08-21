package database

import (
	"database/sql"
	"fmt"

	appConfig "olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"

	_ "github.com/lib/pq"
)

type DB struct {
	conn       *sql.DB
	tableNames map[string]string
}

// jobData represents the job configuration data from database (internal use only)
type jobData struct {
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
func NewDB(cfg *appConfig.Config) (*DB, error) {
	dbURL := getDBURLFromConfig(cfg)

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Apply connection pool settings from config
	conn.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	logger.Debugf("Applied database connection pool settings: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database")

	// Initialize table names based on config
	tableNames := getTableNamesFromConfig(cfg)

	return &DB{
		conn:       conn,
		tableNames: tableNames,
	}, nil
}

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Helper functions
func getDBURLFromConfig(cfg *appConfig.Config) string {
	dbURL := cfg.Database.URL
	if dbURL == "" {
		// Use config values instead of direct env access
		dbURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
			cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)
	}
	return dbURL
}

func getTableNamesFromConfig(cfg *appConfig.Config) map[string]string {
	// Use config RunMode instead of direct env access
	runMode := cfg.Database.RunMode

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
