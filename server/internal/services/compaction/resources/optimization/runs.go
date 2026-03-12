package optimization

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

// CompactionRun represents a single compaction process/run with its metrics
type CompactionRun struct {
	RunID     string         `json:"run-id"`
	Status    string         `json:"status"`
	Type      string         `json:"type"`
	StartTime int64          `json:"start-time"`
	Duration  int64          `json:"duration"`
	Metrics   ProcessMetrics `json:"metrics"`
}

// CompactionRunsResponse represents the response containing list of compaction runs
type CompactionRunsResponse struct {
	Runs  []CompactionRun `json:"runs"`
	Total int             `json:"total"`
}

// returns the list of compaction processes/runs for a particular table along with its metrics
func (s *Service) GetCompactionRuns(ctx context.Context, catalog, database, table string, page, pageSize int) (*CompactionRunsResponse, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", models.APIBase, catalog, database, table)
	params := url.Values{}

	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))

	data, err := s.compaction.Do(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimizing processes for %s.%s.%s: %w", catalog, database, table, err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format for %s.%s.%s", catalog, database, table)
	}

	// Extract list and total
	listData, ok := result["list"].([]interface{})
	if !ok {
		return &CompactionRunsResponse{Runs: []CompactionRun{}, Total: 0}, nil
	}

	total := 0
	if totalVal, ok := result["total"].(float64); ok {
		total = int(totalVal)
	}

	// Parse each run with metrics
	runs := make([]CompactionRun, 0, len(listData))
	for _, item := range listData {
		runMap, _ := item.(map[string]interface{})

		run := CompactionRun{}

		if processID, ok := runMap["processId"].(string); ok {
			run.RunID = processID
		}

		if status, ok := runMap["status"].(string); ok {
			run.Status = status
		}

		if optimizingType, ok := runMap["optimizingType"].(string); ok {
			run.Type = optimizingType
		}

		if startTime, ok := runMap["startTime"].(float64); ok {
			run.StartTime = int64(startTime)
		}

		if duration, ok := runMap["duration"].(float64); ok {
			run.Duration = int64(duration)
		}

		// parse metrics from summary
		if summary, ok := runMap["summary"].(map[string]interface{}); ok {
			run.Metrics = parseProcessMetrics(summary)
		}

		runs = append(runs, run)
	}

	return &CompactionRunsResponse{
		Runs:  runs,
		Total: total,
	}, nil
}

// TODO: cut-down on the metrics required on the UI
// ProcessMetrics represents the metrics for a specific compaction process/run
type ProcessMetrics struct {
	InputDataFilesRewrite      int    `json:"input_data_files_rewrite"`
	InputDataSizeRewrite       string `json:"input_data_size_rewrite"`
	InputDataRecordsRewrite    int64  `json:"input_data_records_rewrite"`
	InputEqualityDeleteFiles   int    `json:"input_equality_delete_files"`
	InputEqualityDeleteSize    string `json:"input_equality_delete_size"`
	InputEqualityDeleteRecords int64  `json:"input_equality_delete_records"`
	InputPositionDeleteFiles   int    `json:"input_position_delete_files,omitempty"`
	InputPositionDeleteSize    string `json:"input_position_delete_size,omitempty"`
	InputPositionDeleteRecords int64  `json:"input_position_delete_records,omitempty"`
	OutputDataFiles            int    `json:"output_data_files"`
	OutputDataSize             string `json:"output_data_size"`
	OutputDataRecords          int64  `json:"output_data_records"`
	OutputDeleteFiles          int    `json:"output_delete_files,omitempty"`
	OutputDeleteSize           string `json:"output_delete_size,omitempty"`
	OutputDeleteRecords        int64  `json:"output_delete_records,omitempty"`
}

// parseProcessMetrics extracts metrics from a process summary map
func parseProcessMetrics(summary map[string]interface{}) ProcessMetrics {
	metrics := ProcessMetrics{}

	if val, ok := summary["input-data-files(rewrite)"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.InputDataFilesRewrite = intVal
		}
	}

	if val, ok := summary["input-data-size(rewrite)"].(string); ok {
		metrics.InputDataSizeRewrite = val
	}

	if val, ok := summary["input-data-records(rewrite)"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.InputDataRecordsRewrite = intVal
		}
	}

	if val, ok := summary["input-equality-delete-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.InputEqualityDeleteFiles = intVal
		}
	}

	if val, ok := summary["input-equality-delete-size"].(string); ok {
		metrics.InputEqualityDeleteSize = val
	}

	if val, ok := summary["input-equality-delete-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.InputEqualityDeleteRecords = intVal
		}
	}

	if val, ok := summary["input-position-delete-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.InputPositionDeleteFiles = intVal
		}
	}

	if val, ok := summary["input-position-delete-size"].(string); ok {
		metrics.InputPositionDeleteSize = val
	}

	if val, ok := summary["input-position-delete-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.InputPositionDeleteRecords = intVal
		}
	}

	if val, ok := summary["output-data-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.OutputDataFiles = intVal
		}
	}

	if val, ok := summary["output-data-size"].(string); ok {
		metrics.OutputDataSize = val
	}

	if val, ok := summary["output-data-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.OutputDataRecords = intVal
		}
	}

	if val, ok := summary["output-delete-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.OutputDeleteFiles = intVal
		}
	}

	if val, ok := summary["output-delete-size"].(string); ok {
		metrics.OutputDeleteSize = val
	}

	if val, ok := summary["output-delete-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.OutputDeleteRecords = intVal
		}
	}

	return metrics
}

func parseInt64FromString(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// CancelCompactionProcess cancels a running compaction process for a table
func (s *Service) CancelCompactionProcess(ctx context.Context, catalog, database, table, processID string) error {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes/%s/cancel",
		models.APIBase, catalog, database, table, processID)

	return s.compaction.DoAndValidate(ctx, http.MethodPost, path, url.Values{}, nil)
}
