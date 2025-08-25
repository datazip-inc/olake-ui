package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	appConfig "github.com/datazip-inc/olake-ui/olake-workers/k8s/config"

	_ "github.com/lib/pq"
)

type DB struct {
	conn       *sql.DB
	tableNames map[string]string
}

// NewDB creates a new database connection
func NewDB(cfg *appConfig.Config) (*DB, error) {
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Optional: apply pooling settings if provided
	if cfg.Database.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime > 0 {
		conn.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	}

	tableNames := map[string]string{
		"job":    fmt.Sprintf("olake-%s-job", cfg.Database.RunMode),
		"source": fmt.Sprintf("olake-%s-source", cfg.Database.RunMode),
		"dest":   fmt.Sprintf("olake-%s-destination", cfg.Database.RunMode),
	}

	return &DB{conn: conn, tableNames: tableNames}, nil
}

// GetJobData retrieves job configuration
func (db *DB) GetJobData(ctx context.Context, jobID int) (map[string]interface{}, error) {
	query := fmt.Sprintf(`
        SELECT j.id, j.name, s.type, s.version, s.config,
               d.dest_type, d.version, d.config, j.streams_config,
               j.state, j.project_id, j.active
        FROM %q j
        JOIN %q s ON j.source_id = s.id
        JOIN %q d ON j.dest_id = d.id
        WHERE j.id = $1
    `, db.tableNames["job"], db.tableNames["source"], db.tableNames["dest"])

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := db.conn.QueryRowContext(cctx, query, jobID)

	var (
		id            int
		name          string
		sourceType    string
		sourceVersion string
		sourceConfig  string
		destType      string
		destVersion   string
		destConfig    string
		streamsConfig string
		state         string
		projectID     string
		active        bool
	)

	if err := row.Scan(&id, &name, &sourceType, &sourceVersion, &sourceConfig,
		&destType, &destVersion, &destConfig, &streamsConfig, &state, &projectID, &active); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":             id,
		"name":           name,
		"source_type":    sourceType,
		"source_version": sourceVersion,
		"source_config":  sourceConfig,
		"dest_type":      destType,
		"dest_version":   destVersion,
		"dest_config":    destConfig,
		"streams_config": streamsConfig,
		"state":          state,
		"project_id":     projectID,
		"active":         active,
	}

	return result, nil
}

// UpdateJobState updates job status
func (db *DB) UpdateJobState(ctx context.Context, jobID int, state string, active bool) error {
	query := fmt.Sprintf(`UPDATE %q SET state = $1, active = $2, updated_at = NOW() WHERE id = $3`, db.tableNames["job"])

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(cctx, query, state, active, jobID)
	return err
}

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.conn.Ping()
}

func (db *DB) Close() error {
	return db.conn.Close()
}
