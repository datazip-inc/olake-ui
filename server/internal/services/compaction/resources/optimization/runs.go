package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

// CompactionRun represents a single compaction process/run
type CompactionRun struct {
	RunID        string `json:"run-id"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	StartTime    int64  `json:"start-time"`
	FinishTime   int64  `json:"finish-time,omitempty"`
	Duration     string `json:"duration"`
	FailReason   string `json:"fail-reason,omitempty"`
	TotalTasks   int    `json:"total-tasks"`
	SuccessTasks int    `json:"success-tasks"`
	RunningTasks int    `json:"running-tasks"`
}

// CompactionRunsResponse represents the response containing list of compaction runs
type CompactionRunsResponse struct {
	Runs  []CompactionRun `json:"runs"`
	Total int             `json:"total"`
}

// GetCompactionRuns fetches the list of compaction processes/runs for a table
func (c *Service) GetCompactionRuns(ctx context.Context, catalog, database, table string, page, pageSize int) (*CompactionRunsResponse, error) {
	// Build the API path
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", models.ApiBase, catalog, database, table)

	// Add query parameters
	params := url.Values{}
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))

	respBody, err := c.compaction.DoRequest(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimizing processes for %s.%s.%s: %w", catalog, database, table, err)
	}

	var resp models.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	// Parse the result
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
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

	// Parse each run
	runs := make([]CompactionRun, 0, len(listData))
	for _, item := range listData {
		runMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		run := CompactionRun{}

		// Extract processId as run-id
		if processID, ok := runMap["processId"].(string); ok {
			run.RunID = processID
		}

		// Extract status
		if status, ok := runMap["status"].(string); ok {
			run.Status = status
		}

		// Extract optimizing type
		if optimizingType, ok := runMap["optimizingType"].(string); ok {
			run.Type = optimizingType
		}

		// Extract start time
		if startTime, ok := runMap["startTime"].(float64); ok {
			run.StartTime = int64(startTime)
		}

		// Extract finish time
		if finishTime, ok := runMap["finishTime"].(float64); ok {
			run.FinishTime = int64(finishTime)
		}

		// Extract duration
		if duration, ok := runMap["duration"].(float64); ok {
			// Convert milliseconds to human-readable format
			durationMs := int64(duration)
			run.Duration = formatDuration(durationMs)
		}

		// Extract fail reason
		if failReason, ok := runMap["failReason"].(string); ok {
			run.FailReason = failReason
		}

		// Extract task counts
		if totalTasks, ok := runMap["totalTasks"].(float64); ok {
			run.TotalTasks = int(totalTasks)
		}
		if successTasks, ok := runMap["successTasks"].(float64); ok {
			run.SuccessTasks = int(successTasks)
		}
		if runningTasks, ok := runMap["runningTasks"].(float64); ok {
			run.RunningTasks = int(runningTasks)
		}

		runs = append(runs, run)
	}

	return &CompactionRunsResponse{
		Runs:  runs,
		Total: total,
	}, nil
}

// formatDuration converts milliseconds to human-readable duration string
func formatDuration(ms int64) string {
	if ms == 0 {
		return "0s"
	}

	seconds := ms / 1000
	minutes := seconds / 60
	hours := minutes / 60

	if hours > 0 {
		remainingMinutes := minutes % 60
		if remainingMinutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		remainingSeconds := seconds % 60
		if remainingSeconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%ds", seconds)
}

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

// ProcessMetricsResponse represents the response containing process metrics
type ProcessMetricsResponse struct {
	Metrics ProcessMetrics `json:"metrics"`
}

// TODO: add pagination
func (c *Service) GetProcessMetrics(ctx context.Context, catalog, database, table, processID string) (*ProcessMetricsResponse, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", models.ApiBase, catalog, database, table)

	params := url.Values{}
	params.Add("page", "1")
	params.Add("pageSize", "10000")

	respBody, err := c.compaction.DoRequest(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimizing processes for %s.%s.%s: %w", catalog, database, table, err)
	}

	var resp models.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	listData, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("process %s not found", processID)
	}

	var processData map[string]interface{}
	for _, item := range listData {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if pid, ok := itemMap["processId"].(string); ok && pid == processID {
			processData = itemMap
			break
		}
	}

	if processData == nil {
		return nil, fmt.Errorf("process %s not found", processID)
	}

	summary, ok := processData["summary"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no summary found for process %s", processID)
	}

	metrics := ProcessMetrics{}

	// Input data files (rewrite)
	if val, ok := summary["input-data-files(rewrite)"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.InputDataFilesRewrite = intVal
		}
	}

	// Input data size (rewrite)
	if val, ok := summary["input-data-size(rewrite)"].(string); ok {
		metrics.InputDataSizeRewrite = val
	}

	// Input data records (rewrite)
	if val, ok := summary["input-data-records(rewrite)"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.InputDataRecordsRewrite = intVal
		}
	}

	// Input equality delete files
	if val, ok := summary["input-equality-delete-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.InputEqualityDeleteFiles = intVal
		}
	}

	// Input equality delete size
	if val, ok := summary["input-equality-delete-size"].(string); ok {
		metrics.InputEqualityDeleteSize = val
	}

	// Input equality delete records
	if val, ok := summary["input-equality-delete-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.InputEqualityDeleteRecords = intVal
		}
	}

	// Input position delete files (optional)
	if val, ok := summary["input-position-delete-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.InputPositionDeleteFiles = intVal
		}
	}

	// Input position delete size (optional)
	if val, ok := summary["input-position-delete-size"].(string); ok {
		metrics.InputPositionDeleteSize = val
	}

	// Input position delete records (optional)
	if val, ok := summary["input-position-delete-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.InputPositionDeleteRecords = intVal
		}
	}

	// Output data files
	if val, ok := summary["output-data-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.OutputDataFiles = intVal
		}
	}

	// Output data size
	if val, ok := summary["output-data-size"].(string); ok {
		metrics.OutputDataSize = val
	}

	// Output data records
	if val, ok := summary["output-data-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.OutputDataRecords = intVal
		}
	}

	// Output delete files (optional)
	if val, ok := summary["output-delete-files"].(string); ok {
		if intVal, err := parseIntFromString(val); err == nil {
			metrics.OutputDeleteFiles = intVal
		}
	}

	// Output delete size (optional)
	if val, ok := summary["output-delete-size"].(string); ok {
		metrics.OutputDeleteSize = val
	}

	// Output delete records (optional)
	if val, ok := summary["output-delete-records"].(string); ok {
		if intVal, err := parseInt64FromString(val); err == nil {
			metrics.OutputDeleteRecords = intVal
		}
	}

	return &ProcessMetricsResponse{
		Metrics: metrics,
	}, nil
}


func parseInt64FromString(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
