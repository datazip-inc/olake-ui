package database

import (
	"encoding/json"
	"olake-ui/olake-workers/k8s/utils/parser"
)

// ParseJobOutput extracts meaningful information from job output logs
func ParseJobOutput(output string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Try to find JSON in the output (similar to Docker implementation)
	lines := parser.SplitLines(output)

	for _, line := range lines {
		// Try to parse each line as JSON
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
			// Found valid JSON, merge with result
			for k, v := range jsonData {
				result[k] = v
			}
		}
	}

	// If no JSON found, return basic info
	if len(result) == 0 {
		result["raw_output"] = output
		result["status"] = "completed"
	}

	return result, nil
}

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
