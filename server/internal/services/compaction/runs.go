package compaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ProcessMetrics represents the metrics for a specific compaction process/run
type ProcessMetrics struct {
	InputDataFilesRewrite       int    `json:"input_data_files_rewrite"`
	InputDataSizeRewrite        string `json:"input_data_size_rewrite"`
	InputDataRecordsRewrite     int64  `json:"input_data_records_rewrite"`
	InputEqualityDeleteFiles    int    `json:"input_equality_delete_files"`
	InputEqualityDeleteSize     string `json:"input_equality_delete_size"`
	InputEqualityDeleteRecords  int64  `json:"input_equality_delete_records"`
	InputPositionDeleteFiles    int    `json:"input_position_delete_files,omitempty"`
	InputPositionDeleteSize     string `json:"input_position_delete_size,omitempty"`
	InputPositionDeleteRecords  int64  `json:"input_position_delete_records,omitempty"`
	OutputDataFiles             int    `json:"output_data_files"`
	OutputDataSize              string `json:"output_data_size"`
	OutputDataRecords           int64  `json:"output_data_records"`
	OutputDeleteFiles           int    `json:"output_delete_files,omitempty"`
	OutputDeleteSize            string `json:"output_delete_size,omitempty"`
	OutputDeleteRecords         int64  `json:"output_delete_records,omitempty"`
}

// ProcessMetricsResponse represents the response containing process metrics
type ProcessMetricsResponse struct {
	Metrics ProcessMetrics `json:"metrics"`
}

// GetProcessMetrics fetches the metrics for a specific compaction process
func (c *Compaction) GetProcessMetrics(ctx context.Context, catalog, database, table, processID string) (*ProcessMetricsResponse, error) {
	// Build the API path to get the specific process
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", apiBase, catalog, database, table)
	
	// We need to fetch all processes and filter by processId
	// Alternatively, we can use a larger page size to ensure we get the process
	params := url.Values{}
	params.Add("page", "1")
	params.Add("pageSize", "1000") // Large enough to find the process
	
	respBody, err := c.doRequest(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimizing processes for %s.%s.%s: %w", catalog, database, table, err)
	}

	var resp Response
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

	// Extract list
	listData, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("process %s not found", processID)
	}

	// Find the specific process by processId
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

	// Extract summary which contains the metrics
	summary, ok := processData["summary"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no summary found for process %s", processID)
	}

	// Parse metrics from summary
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

// Helper function to parse int64 from string
func parseInt64FromString(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
