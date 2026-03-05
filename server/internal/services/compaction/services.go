package compaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/database"
)

type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// GetCatalogs returns the list of catalogs from fusion
func (c *Compaction) GetCatalogs(ctx context.Context) (interface{}, error) {
	path := apiBase + "catalogs"
	respBody, err := c.doRequest(ctx, http.MethodGet, path, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all catalogs: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	// Result can be an array or object, so unmarshal to interface{}
	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse catalog result: %w", err)
	}

	return result, nil
}

// GetDatabases returns the list of databases for a given catalog
func (c *Compaction) GetDatabases(ctx context.Context, catalog string, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases", apiBase, catalog)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	respBody, err := c.doRequest(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get databases for catalog %s: %w", catalog, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse databases result: %w", err)
	}

	return result, nil
}

// GetTables returns the list of tables for a given catalog and database
func (c *Compaction) GetTables(ctx context.Context, catalog string, database string, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases/%s/tables", apiBase, catalog, database)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	respBody, err := c.doRequest(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, database, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tables result: %w", err)
	}

	return result, nil
}

// GetTableDetails returns the details of a specific table including size information
func (c *Compaction) GetTableDetails(ctx context.Context, catalog string, database string, table string) (interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/details", apiBase, catalog, database, table)

	params := url.Values{}
	params.Set("token", "")

	respBody, err := c.doRequest(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get table details for %s.%s.%s: %w", catalog, database, table, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse table details result: %w", err)
	}

	return result, nil
}

// GetOptimizingProcesses returns the optimization process history for a table
func (c *Compaction) GetOptimizingProcesses(ctx context.Context, catalog string, database string, table string) (interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", apiBase, catalog, database, table)

	params := url.Values{}

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

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse optimizing processes result: %w", err)
	}

	return result, nil
}

// DashboardResponse represents the complete dashboard data structure
type DashboardResponse struct {
	Catalogs []CatalogWithData `json:"catalogs"`
}

type CatalogWithData struct {
	Name      string             `json:"name"`
	Type      string             `json:"type,omitempty"`
	Databases []DatabaseWithData `json:"databases"`
}

type DatabaseWithData struct {
	Name   string      `json:"name"`
	Tables []TableInfo `json:"tables"`
}

// CatalogsResponse represents catalogs with their databases (no table details)
type CatalogsResponse struct {
	Catalogs []CatalogWithDatabases `json:"catalogs"`
}

type CatalogWithDatabases struct {
	Name      string   `json:"name"`
	Type      string   `json:"type,omitempty"`
	Databases []string `json:"databases"`
}

// TablesResponse represents tables with full details for a specific catalog/database
type TablesResponse struct {
	Catalog  string      `json:"catalog"`
	Database string      `json:"database"`
	Tables   []TableInfo `json:"tables"`
}

type TableInfo struct {
	Name      string            `json:"name"`
	TotalSize string            `json:"totalSize,omitempty"`
	ByOLake   bool              `json:"byOLake"`
	Major     *OptimizationInfo `json:"major"`
	Minor     *OptimizationInfo `json:"minor"`
	Full      *OptimizationInfo `json:"full"`
	Enabled   bool              `json:"enabled"`
}

type OptimizationInfo struct {
	LastRun string `json:"last-run,omitempty"`
	Status  string `json:"status,omitempty"`
}

// GetCatalogsWithDatabases fetches all catalogs and their databases (without table details)
func (c *Compaction) GetCatalogsWithDatabases(ctx context.Context) (*CatalogsResponse, error) {
	// Step 1: Get all catalogs
	catalogsResult, err := c.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs: %w", err)
	}

	// Parse catalogs result
	catalogsList, ok := catalogsResult.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected catalogs format, got type: %T", catalogsResult)
	}

	response := &CatalogsResponse{
		Catalogs: make([]CatalogWithDatabases, 0, len(catalogsList)),
	}

	// Step 2: For each catalog, get databases
	for _, catalogItem := range catalogsList {
		catalogMap, ok := catalogItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Amoro uses "catalogName" and "catalogType" fields
		catalogName, ok := catalogMap["catalogName"].(string)
		if !ok {
			if nameVal := catalogMap["catalogName"]; nameVal != nil {
				catalogName = fmt.Sprintf("%v", nameVal)
			}
		}

		catalogType, ok := catalogMap["catalogType"].(string)
		if !ok {
			if typeVal := catalogMap["catalogType"]; typeVal != nil {
				catalogType = fmt.Sprintf("%v", typeVal)
			}
		}

		if catalogName == "" {
			continue
		}

		catalogData := CatalogWithDatabases{
			Name:      catalogName,
			Type:      catalogType,
			Databases: make([]string, 0),
		}

		// Get databases for this catalog
		databasesResult, err := c.GetDatabases(ctx, catalogName, "")
		if err != nil {
			fmt.Printf("Failed to get databases for catalog %s: %v\n", catalogName, err)
			response.Catalogs = append(response.Catalogs, catalogData)
			continue
		}

		databasesList, ok := databasesResult.([]interface{})
		if !ok {
			response.Catalogs = append(response.Catalogs, catalogData)
			continue
		}

		// Extract database names
		for _, dbItem := range databasesList {
			dbName, ok := dbItem.(string)
			if ok && dbName != "" {
				catalogData.Databases = append(catalogData.Databases, dbName)
			}
		}

		response.Catalogs = append(response.Catalogs, catalogData)
	}

	return response, nil
}

// GetTablesWithDetails fetches all tables with full details for a specific catalog and database
func (c *Compaction) GetTablesWithDetails(ctx context.Context, catalog string, databaseName string, db *database.Database) (*TablesResponse, error) {
	response := &TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   make([]TableInfo, 0),
	}

	// Get tables for this database
	tablesResult, err := c.GetTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, databaseName, err)
	}

	tablesList, ok := tablesResult.([]interface{})
	if !ok {
		return response, nil
	}

	// Parse tables and fetch details for each
	for _, tableItem := range tablesList {
		tableMap, ok := tableItem.(map[string]interface{})
		if !ok {
			continue
		}

		tableName, _ := tableMap["name"].(string)

		if tableName != "" {
			tableInfo := TableInfo{
				Name:    tableName,
				ByOLake: false,
				Enabled: false,
			}

			// Check if table is managed by OLake
			if db != nil {
				isManagedByOLake, err := db.CheckTableManagedByOLake(catalog, databaseName, tableName)
				if err != nil {
					fmt.Printf("Failed to check if table %s.%s.%s is managed by OLake: %v\n", catalog, databaseName, tableName, err)
				} else {
					tableInfo.ByOLake = isManagedByOLake
				}
			}

			// Fetch table details to get totalSize
			tableDetails, err := c.GetTableDetails(ctx, catalog, databaseName, tableName)
			if err != nil {
				fmt.Printf("Failed to get details for table %s.%s.%s: %v\n", catalog, databaseName, tableName, err)
			} else {
				// Extract totalSize from baseMetrics
				if detailsMap, ok := tableDetails.(map[string]interface{}); ok {
					if baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{}); ok {
						if totalSize, ok := baseMetrics["totalSize"].(string); ok {
							tableInfo.TotalSize = totalSize
						}
					}
				}
			}

			// Fetch optimizing processes to get MAJOR, MINOR, FULL optimization data
			optimizingProcesses, err := c.GetOptimizingProcesses(ctx, catalog, databaseName, tableName)
			if err != nil {
				fmt.Printf("Failed to get optimizing processes for table %s.%s.%s: %v\n", catalog, databaseName, tableName, err)
			} else {
				// Parse optimization processes
				if processesMap, ok := optimizingProcesses.(map[string]interface{}); ok {
					if processList, ok := processesMap["list"].([]interface{}); ok {
						// Track latest process for each type
						latestProcesses := make(map[string]map[string]interface{})

						// Iterate through all processes to find the latest for each type
						for _, processItem := range processList {
							if process, ok := processItem.(map[string]interface{}); ok {
								optimizingType, _ := process["optimizingType"].(string)
								finishTime, _ := process["finishTime"].(float64)

								if optimizingType != "" {
									// Check if this is the latest process for this type
									if existing, exists := latestProcesses[optimizingType]; !exists {
										latestProcesses[optimizingType] = process
									} else {
										existingFinishTime, _ := existing["finishTime"].(float64)
										if finishTime > existingFinishTime {
											latestProcesses[optimizingType] = process
										}
									}
								}
							}
						}

						// Extract data for MAJOR, MINOR, FULL
						for processType, process := range latestProcesses {
							finishTime, _ := process["finishTime"].(float64)
							status, _ := process["status"].(string)

							// Convert timestamp to relative time format
							var lastRun string
							if finishTime > 0 {
								timestamp := time.Unix(0, int64(finishTime)*int64(time.Millisecond))
								duration := time.Since(timestamp)

								if duration < time.Minute {
									seconds := int(duration.Seconds())
									if seconds == 1 {
										lastRun = "1 sec ago"
									} else {
										lastRun = fmt.Sprintf("%d secs ago", seconds)
									}
								} else if duration < time.Hour {
									minutes := int(duration.Minutes())
									if minutes == 1 {
										lastRun = "1 minute ago"
									} else {
										lastRun = fmt.Sprintf("%d minutes ago", minutes)
									}
								} else {
									hours := int(duration.Hours())
									if hours == 1 {
										lastRun = "1 hour ago"
									} else {
										lastRun = fmt.Sprintf("%d hours ago", hours)
									}
								}
							}

							optimizationInfo := &OptimizationInfo{
								LastRun: lastRun,
								Status:  status,
							}

							switch processType {
							case "MAJOR":
								tableInfo.Major = optimizationInfo
							case "MINOR":
								tableInfo.Minor = optimizationInfo
							case "FULL":
								tableInfo.Full = optimizationInfo
							}
						}
					}
				}
			}

			response.Tables = append(response.Tables, tableInfo)
		}
	}

	return response, nil
}

// SetTablePropertiesRequest represents the request to set table properties
type SetTablePropertiesRequest struct {
	Catalog    string            `json:"catalog"`
	Database   string            `json:"database"`
	Table      string            `json:"table"`
	Properties map[string]string `json:"properties"`
}

// SetTablePropertiesResponse represents the response from setting table properties
type SetTablePropertiesResponse struct {
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}

// TerminalExecuteRequest represents the request body for terminal SQL execution
type TerminalExecuteRequest struct {
	SQL string `json:"sql"`
}

// TerminalSessionResponse represents the response from terminal execute
type TerminalSessionResponse struct {
	SessionID string `json:"sessionId"`
}

// SQLResult represents the result of a SQL execution
type SQLResult struct {
	ID      string     `json:"id"`
	Status  string     `json:"status"`
	Columns []string   `json:"columns,omitempty"`
	RowData [][]string `json:"rowData,omitempty"`
}

// SetTableProperties sets table properties using the Terminal API (for external catalogs)
// This method uses ALTER TABLE SET TBLPROPERTIES SQL statement
func (c *Compaction) SetTableProperties(ctx context.Context, req SetTablePropertiesRequest) (*SetTablePropertiesResponse, error) {
	// Build ALTER TABLE SQL statement
	var propsSQL []string
	for key, value := range req.Properties {
		propsSQL = append(propsSQL, fmt.Sprintf("'%s' = '%s'", key, value))
	}

	sql := fmt.Sprintf(
		"ALTER TABLE %s.%s SET TBLPROPERTIES (%s)",
		req.Database,
		req.Table,
		strings.Join(propsSQL, ", "),
	)

	// Execute via Terminal API
	path := fmt.Sprintf("%sterminal/catalogs/%s/execute", apiBase, req.Catalog)

	requestBody := TerminalExecuteRequest{
		SQL: sql,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, path, url.Values{}, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ALTER TABLE for %s.%s.%s: %w", req.Catalog, req.Database, req.Table, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	// Parse session ID from result
	var sessionResult TerminalSessionResponse
	if err := json.Unmarshal(resp.Result, &sessionResult); err != nil {
		return nil, fmt.Errorf("failed to parse session result: %w", err)
	}

	// Poll for SQL completion
	status, err := c.waitForSQLCompletion(ctx, sessionResult.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for SQL completion: %w", err)
	}

	// Determine success based on status
	success := status == "FINISHED"
	message := fmt.Sprintf("ALTER TABLE command %s. Session ID: %s", status, sessionResult.SessionID)

	if status == "FAILED" {
		message = fmt.Sprintf("ALTER TABLE command failed. Session ID: %s", sessionResult.SessionID)
	}

	return &SetTablePropertiesResponse{
		SessionID: sessionResult.SessionID,
		Status:    status,
		Success:   success,
		Message:   message,
	}, nil
}

// waitForSQLCompletion polls the terminal API until SQL execution completes
func (c *Compaction) waitForSQLCompletion(ctx context.Context, sessionID string) (string, error) {
	maxAttempts := 60 // Poll for up to 60 seconds
	pollInterval := 1 * time.Second

	for i := 0; i < maxAttempts; i++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return "CANCELED", ctx.Err()
		default:
		}

		// Get SQL result status
		path := fmt.Sprintf("%sterminal/%s/result", apiBase, sessionID)
		respBody, err := c.doRequest(ctx, http.MethodGet, path, url.Values{}, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get SQL result: %w", err)
		}

		var resp Response
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		if resp.Code != 200 && resp.Code != 0 {
			return "", fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
		}

		// Parse SQL results
		var results []SQLResult
		if err := json.Unmarshal(resp.Result, &results); err != nil {
			return "", fmt.Errorf("failed to parse SQL results: %w", err)
		}

		// Check if we have results
		if len(results) > 0 {
			status := results[0].Status

			// Check if execution is complete
			if status == "FINISHED" || status == "FAILED" || status == "CANCELED" {
				return status, nil
			}
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}

	// Timeout reached
	return "TIMEOUT", fmt.Errorf("SQL execution timed out after %d seconds", maxAttempts)
}

// EnableSelfOptimizing enables self-optimizing for a table
func (c *Compaction) EnableSelfOptimizing(ctx context.Context, catalog, database, table string) (*SetTablePropertiesResponse, error) {
	return c.SetTableProperties(ctx, SetTablePropertiesRequest{
		Catalog:  catalog,
		Database: database,
		Table:    table,
		Properties: map[string]string{
			"self-optimizing.enabled": "true",
		},
	})
}

// DisableSelfOptimizing disables self-optimizing for a table
func (c *Compaction) DisableSelfOptimizing(ctx context.Context, catalog, database, table string) (*SetTablePropertiesResponse, error) {
	return c.SetTableProperties(ctx, SetTablePropertiesRequest{
		Catalog:  catalog,
		Database: database,
		Table:    table,
		Properties: map[string]string{
			"self-optimizing.enabled": "false",
		},
	})
}

// TableMetricsResponse represents detailed metrics for a table
type TableMetricsResponse struct {
	TableMetrics TableMetrics `json:"table-metrics"`
}

type TableMetrics struct {
	FileCount       FileCount `json:"file-count"`
	AverageFileSize string    `json:"average-file-size"`
	LastCommitTime  int64     `json:"last-commit-time,omitempty"`
}

type FileCount struct {
	Total       int         `json:"total"`
	DataFiles   int         `json:"data-files"`
	DeleteFiles DeleteFiles `json:"delete-files"`
}

type DeleteFiles struct {
	Equality   int `json:"equality"`
	Positional int `json:"positional"`
}

// GetTableMetrics fetches detailed file metrics for a specific table
func (c *Compaction) GetTableMetrics(ctx context.Context, catalog, database, table string) (*TableMetricsResponse, error) {
	// Get table details which contains baseMetrics
	tableDetails, err := c.GetTableDetails(ctx, catalog, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table details for %s.%s.%s: %w", catalog, database, table, err)
	}

	response := &TableMetricsResponse{
		TableMetrics: TableMetrics{
			FileCount: FileCount{
				DeleteFiles: DeleteFiles{
					Equality:   0,
					Positional: 0,
				},
			},
		},
	}

	// Parse table details to extract metrics
	if detailsMap, ok := tableDetails.(map[string]interface{}); ok {
		if baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{}); ok {
			// Extract file count
			if fileCount, ok := baseMetrics["fileCount"].(float64); ok {
				response.TableMetrics.FileCount.Total = int(fileCount)
			}

			// Extract average file size
			if avgSize, ok := baseMetrics["averageFileSize"].(string); ok {
				response.TableMetrics.AverageFileSize = avgSize
			}

			// Extract last commit time
			if lastCommitTime, ok := baseMetrics["lastCommitTime"].(float64); ok {
				response.TableMetrics.LastCommitTime = int64(lastCommitTime)
			}
		}

		// Try to get detailed file breakdown from tableSummary
		if tableSummary, ok := detailsMap["tableSummary"].(map[string]interface{}); ok {
			if summary, ok := tableSummary["summary"].(map[string]interface{}); ok {
				// Extract data files count
				if dataFiles, ok := summary["total-data-files"].(string); ok {
					if count, err := parseIntFromString(dataFiles); err == nil {
						response.TableMetrics.FileCount.DataFiles = count
					}
				} else if dataFiles, ok := summary["total-data-files"].(float64); ok {
					response.TableMetrics.FileCount.DataFiles = int(dataFiles)
				}

				// Extract delete files count
				if deleteFiles, ok := summary["total-delete-files"].(string); ok {
					if count, err := parseIntFromString(deleteFiles); err == nil {
						totalDeleteFiles := count
						// Try to get equality and positional breakdown
						if eqDeletes, ok := summary["total-equality-deletes"].(string); ok {
							if eqCount, err := parseIntFromString(eqDeletes); err == nil {
								response.TableMetrics.FileCount.DeleteFiles.Equality = eqCount
							}
						} else if eqDeletes, ok := summary["total-equality-deletes"].(float64); ok {
							response.TableMetrics.FileCount.DeleteFiles.Equality = int(eqDeletes)
						}

						if posDeletes, ok := summary["total-positional-deletes"].(string); ok {
							if posCount, err := parseIntFromString(posDeletes); err == nil {
								response.TableMetrics.FileCount.DeleteFiles.Positional = posCount
							}
						} else if posDeletes, ok := summary["total-positional-deletes"].(float64); ok {
							response.TableMetrics.FileCount.DeleteFiles.Positional = int(posDeletes)
						}

						// If we don't have the breakdown, put all delete files in equality
						if response.TableMetrics.FileCount.DeleteFiles.Equality == 0 &&
						   response.TableMetrics.FileCount.DeleteFiles.Positional == 0 {
							response.TableMetrics.FileCount.DeleteFiles.Equality = totalDeleteFiles
						}
					}
				} else if deleteFiles, ok := summary["total-delete-files"].(float64); ok {
					response.TableMetrics.FileCount.DeleteFiles.Equality = int(deleteFiles)
				}
			}
		}

		// If we couldn't get data files from summary, calculate from total - delete files
		if response.TableMetrics.FileCount.DataFiles == 0 && response.TableMetrics.FileCount.Total > 0 {
			totalDeleteFiles := response.TableMetrics.FileCount.DeleteFiles.Equality +
			                   response.TableMetrics.FileCount.DeleteFiles.Positional
			response.TableMetrics.FileCount.DataFiles = response.TableMetrics.FileCount.Total - totalDeleteFiles
		}
	}

	return response, nil
}

// Helper function to parse integer from string
func parseIntFromString(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// CompactionRun represents a single compaction process/run
type CompactionRun struct {
	RunID       string `json:"run-id"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	StartTime   int64  `json:"start-time"`
	FinishTime  int64  `json:"finish-time,omitempty"`
	Duration    string `json:"duration"`
	FailReason  string `json:"fail-reason,omitempty"`
	TotalTasks  int    `json:"total-tasks"`
	SuccessTasks int   `json:"success-tasks"`
	RunningTasks int   `json:"running-tasks"`
}

// CompactionRunsResponse represents the response containing list of compaction runs
type CompactionRunsResponse struct {
	Runs  []CompactionRun `json:"runs"`
	Total int             `json:"total"`
}

// GetCompactionRuns fetches the list of compaction processes/runs for a table
func (c *Compaction) GetCompactionRuns(ctx context.Context, catalog, database, table string, page, pageSize int) (*CompactionRunsResponse, error) {
	// Build the API path
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", apiBase, catalog, database, table)

	// Add query parameters
	params := url.Values{}
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))

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
