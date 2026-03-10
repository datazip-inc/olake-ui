package database

import (
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

type StreamsConfig struct {
	SelectedStreams map[string][]StreamInfo `json:"selected_streams"`
}

type StreamInfo struct {
	PartitionRegex string `json:"partition_regex"`
	StreamName     string `json:"stream_name"`
	Normalization  bool   `json:"normalization"`
}

func (db *Database) CheckTableManagedByOLake(catalogName, databaseName, tableName string) (bool, error) {
	var jobs []*models.Job

	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.JobTable]).
		RelatedSel("DestID").
		All(&jobs)

	if err != nil {
		return false, fmt.Errorf("failed to query jobs: %w", err)
	}

	for _, job := range jobs {
		if job.DestID == nil || job.DestID.Name != catalogName {
			continue
		}

		var streamsConfig StreamsConfig
		if err := json.Unmarshal([]byte(job.StreamsConfig), &streamsConfig); err != nil {
			continue
		}

		if streams, exists := streamsConfig.SelectedStreams[databaseName]; exists {
			for _, stream := range streams {
				if stream.StreamName == tableName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
