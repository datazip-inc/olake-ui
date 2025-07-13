package database

func (db *DB) GetJobData(jobID int) (*JobData, error) {
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

	var jobData JobData
	err := db.conn.QueryRow(query, jobID).Scan(
		&jobData.ID, &jobData.Name, &jobData.Active,
		&jobData.SourceType, &jobData.SourceVersion, &jobData.SourceConfig,
		&jobData.DestType, &jobData.DestConfig,
		&jobData.StreamsConfig, &jobData.State,
	)

	return &jobData, err
}

// UpdateJobState updates the job state and active status in the database
func (db *DB) UpdateJobState(jobID int, state string, active bool) error {
	query := `UPDATE "olake-dev-job" SET state = $1, active = $2, updated_at = NOW() WHERE id = $3`
	_, err := db.conn.Exec(query, state, active, jobID)
	return err
}
