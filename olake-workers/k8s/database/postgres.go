package database

import (
	"database/sql"
	"fmt"
	"os"

	"olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/utils/env"

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
func NewDB() (*DB, error) {
	dbURL := getDBURL()

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Apply connection pool settings
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warnf("Failed to load database config, using defaults: %v", err)
	} else {
		conn.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		conn.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
		logger.Infof("Applied database connection pool settings: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v", 
			cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)
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

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.conn.Ping()
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
		host := env.GetEnv("DB_HOST", "localhost")
		port := env.GetEnv("DB_PORT", "5432")
		user := env.GetEnv("DB_USER", "olake")
		password := env.GetEnv("DB_PASSWORD", "password")
		dbname := env.GetEnv("DB_NAME", "olake")
		sslmode := env.GetEnv("DB_SSLMODE", "disable")

		dbURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}
	return dbURL
}

func getTableNames() map[string]string {
	// Get environment mode (development, production, etc.)
	runMode := env.GetEnv("RUN_MODE", "dev")

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
