package database

import (
	"context"
	"fmt"
	"time"
)

// GetJobData retrieves job and related config from DB using context and tableNames
func (db *DB) GetJobData(ctx context.Context, jobID int) (*jobData, error) {
	jobTable := db.tableNames["job"]
	sourceTable := db.tableNames["source"]
	destTable := db.tableNames["destination"]

	query := fmt.Sprintf(`
        SELECT
            j.id,
            j.name,
            s.type,
            s.version,
            s.config,
            d.dest_type,
            d.version,
            d.config,
            j.streams_config,
            j.state,
            j.project_id,
            j.active
        FROM %q j
        JOIN %q s ON j.source_id = s.id
        JOIN %q d ON j.dest_id = d.id
        WHERE j.id = $1
    `, jobTable, sourceTable, destTable)

	// Add a timeout on DB call
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	var result jobData
	err := db.conn.QueryRowContext(ctx, query, jobID).Scan(
		&result.ID,
		&result.Name,
		&result.SourceType,
		&result.SourceVersion,
		&result.SourceConfig,
		&result.DestType,
		&result.DestVersion,
		&result.DestConfig,
		&result.StreamsConfig,
		&result.State,
		&result.ProjectID,
		&result.Active,
	)
	return &result, err
}

// UpdateJobState updates the job state and active status in the database using context
func (db *DB) UpdateJobState(ctx context.Context, jobID int, state string, active bool) error {
	jobTable := db.tableNames["job"]
	query := fmt.Sprintf(`UPDATE %q SET state = $1, active = $2, updated_at = NOW() WHERE id = $3`, jobTable)

	// Add a timeout on DB call
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	_, err := db.conn.ExecContext(ctx, query, state, active, jobID)
	return err
}
