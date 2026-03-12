package database

import (
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

type StreamsConfig struct {
	SelectedStreams map[string][]SelectedStream `json:"selected_streams"`
	Streams         []StreamWrapper             `json:"streams"`
}

type SelectedStream struct {
	StreamName string `json:"stream_name"`
}

type StreamWrapper struct {
	Stream Stream `json:"stream"`
}

type Stream struct {
	Name                string `json:"name"`
	DestinationDatabase string `json:"destination_database"`
	DestinationTable    string `json:"destination_table"`
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
		// destination_name == catalogName in compaction
		if job.DestID.Name != catalogName {
			continue
		}

		var streamsConfig StreamsConfig
		if err := json.Unmarshal([]byte(job.StreamsConfig), &streamsConfig); err != nil {
			continue
		}

		selectedStreams, exists := streamsConfig.SelectedStreams[databaseName]
		if !exists {
			continue
		}

		selectedStreamNames := make(map[string]bool)
		for _, selected := range selectedStreams {
			selectedStreamNames[selected.StreamName] = true
		}

		for _, streamWrapper := range streamsConfig.Streams {
			stream := streamWrapper.Stream
			if selectedStreamNames[stream.Name] &&
				stream.DestinationDatabase == databaseName &&
				stream.DestinationTable == tableName {
				return true, nil
			}
		}
	}

	return false, nil
}
