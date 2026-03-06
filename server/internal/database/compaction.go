package database

import (
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// StreamsConfig represents the structure of the streams_config JSON field
type StreamsConfig struct {
	SelectedStreams map[string][]StreamInfo `json:"selected_streams"`
}

type StreamInfo struct {
	PartitionRegex string `json:"partition_regex"`
	StreamName     string `json:"stream_name"`
	Normalization  bool   `json:"normalization"`
}

// CheckTableManagedByOLake checks if a table (stream) exists in any job's streams_config
// for the given catalog (destination name) and database
func (db *Database) CheckTableManagedByOLake(catalogName string, databaseName string, tableName string) (bool, error) {
	// First, find all jobs where the destination name matches the catalog name
	var jobs []*models.Job
	
	// Query jobs with related destination
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).
		RelatedSel("DestID").
		All(&jobs)
	
	if err != nil {
		return false, fmt.Errorf("failed to query jobs: %w", err)
	}

	// Check each job's streams_config
	for _, job := range jobs {
		// Skip if destination is nil or name doesn't match catalog
		if job.DestID == nil || job.DestID.Name != catalogName {
			continue
		}

		// Parse the streams_config JSON
		var streamsConfig StreamsConfig
		if err := json.Unmarshal([]byte(job.StreamsConfig), &streamsConfig); err != nil {
			// Skip jobs with invalid JSON
			continue
		}

		// Check if the database exists in selected_streams
		if streams, exists := streamsConfig.SelectedStreams[databaseName]; exists {
			// Check if the table name exists in the streams list
			for _, stream := range streams {
				if stream.StreamName == tableName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
