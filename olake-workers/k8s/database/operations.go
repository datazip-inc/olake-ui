package database

func (db *DB) GetJobData(jobID int) (*jobData, error) {
	query := `
        SELECT 
            j.id, j.name, j.active,
            s.type, s.version, s.config,
            d.dest_type, d.config,
            j.streams_config, j.state
        FROM "olake-dev-job" j
        JOIN "olake-dev-source" s ON j.source_id = s.id  
        JOIN "olake-dev-destination" d ON j.dest_id = d.id
        WHERE j.id = $1`

	var result jobData
	err := db.conn.QueryRow(query, jobID).Scan(
		&result.ID, &result.Name, &result.Active,
		&result.SourceType, &result.SourceVersion, &result.SourceConfig,
		&result.DestType, &result.DestConfig,
		&result.StreamsConfig, &result.State,
	)

	return &result, err
}

// UpdateJobState updates the job state and active status in the database
func (db *DB) UpdateJobState(jobID int, state string, active bool) error {
	query := `UPDATE "olake-dev-job" SET state = $1, active = $2, updated_at = NOW() WHERE id = $3`
	_, err := db.conn.Exec(query, state, active, jobID)
	return err
}
