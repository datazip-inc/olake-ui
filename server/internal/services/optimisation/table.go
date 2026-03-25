package optimisation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// GetTablesWithDetails fetches all tables with full details for a specific catalog and database
func (s *Service) GetTablesWithDetails(ctx context.Context, catalog, databaseName string) (*dto.TablesResponse, error) {
	response := &dto.TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   make([]dto.TableInfo, 0),
	}

	tablesResult, err := s.GetTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %s", catalog, databaseName, err)
	}

	tablesList, ok := tablesResult.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected tables result format for %s.%s: got %T", catalog, databaseName, tablesResult)
	}

	for _, item := range tablesList {
		tableMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid table item type: got %T", item)
		}

		tableName, ok := tableMap["name"].(string)
		if !ok || tableName == "" {
			return nil, fmt.Errorf("missing or invalid table name in %v", tableMap)
		}

		tableInfo := dto.TableInfo{
			Name:    tableName,
			Enabled: false,
			ByOLake: false,
		}

		tableDetails, err := s.GetTableDetails(ctx, catalog, databaseName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get details for table %s.%s.%s: %s", catalog, databaseName, tableName, err)
		}

		detailsMap, ok := tableDetails.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid tableDetails type: expected map[string]interface{}, got %T", tableDetails)
		}

		baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid baseMetrics for table %s.%s.%s", catalog, databaseName, tableName)
		}

		totalSize, ok := baseMetrics["totalSize"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid totalSize in baseMetrics for table %s.%s.%s", catalog, databaseName, tableName)
		}

		tableInfo.TotalSize = totalSize

		properties, ok := detailsMap["properties"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid properties for table %s.%s.%s", catalog, databaseName, tableName)
		}

		if enabled, ok := properties["self-optimizing.enabled"]; ok {
			switch v := enabled.(type) {
			case string:
				tableInfo.Enabled = v == "true"
			case bool:
				tableInfo.Enabled = v
			default:
				tableInfo.Enabled = fmt.Sprint(v) == "true"
			}
		}

		if _, ok := properties["olake-2pc"]; ok {
			tableInfo.ByOLake = true
		}

		tableSummary, ok := detailsMap["tableSummary"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid tableSummary for table %s.%s.%s", catalog, databaseName, tableName)
		}

		healthScore, ok := tableSummary["healthScore"].(float64)
		if !ok {
			return nil, fmt.Errorf("missing or invalid healthScore in tableSummary for table %s.%s.%s", catalog, databaseName, tableName)
		}

		tableInfo.HealthScore = int(healthScore)

		// fetch latest optimizing processes for each type
		res, err := s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MINOR")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest minor process info: %s", err)
		}
		tableInfo.Minor = res

		res, err = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MAJOR")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest major process info: %s", err)
		}
		tableInfo.Major = res

		res, err = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "FULL")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest full process info: %s", err)
		}
		tableInfo.Full = res

		response.Tables = append(response.Tables, tableInfo)
	}

	return response, nil
}

// GetTables returns the list of tables for a given catalog and database
func (s *Service) GetTables(ctx context.Context, catalog, database, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases/%s/tables", constants.OptimisationAPIBase, catalog, database)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	return s.Do(ctx, http.MethodGet, path, params, nil)
}

// returns the details of a specific table including size information
func (s *Service) GetTableDetails(ctx context.Context, catalog, database, table string) (interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/details", constants.OptimisationAPIBase, catalog, database, table)

	return s.Do(ctx, http.MethodGet, path, url.Values{}, nil)
}

// fetchLatestProcessInfo fetches the latest optimizing process info for a specific type
func (s *Service) fetchLatestProcessInfo(ctx context.Context, catalog, database, table, processType string) (*dto.OptimizationInfo, error) {
	result, err := s.GetLatestOptimizingProcessByType(ctx, catalog, database, table, processType)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest %s optimizing process for %s.%s.%s: %s", processType, catalog, database, table, err)
	}

	processList, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to fetch optimisation process info")
	}

	if len(processList) == 0 {
		return nil, nil
	}

	// Get the first (latest) process
	process, ok := processList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid process format")
	}

	finishTime, _ := process["finishTime"].(float64)
	status, _ := process["status"].(string)
	runID, _ := process["processId"].(string)

	var lastRun string
	if finishTime > 0 {
		timestamp := time.Unix(0, int64(finishTime)*int64(time.Millisecond))
		duration := time.Since(timestamp)

		switch {
		case duration < time.Minute:
			seconds := int(duration.Seconds())
			if seconds == 1 {
				lastRun = "1 sec ago"
			} else {
				lastRun = fmt.Sprintf("%d secs ago", seconds)
			}
		case duration < time.Hour:
			minutes := int(duration.Minutes())
			if minutes == 1 {
				lastRun = "1 minute ago"
			} else {
				lastRun = fmt.Sprintf("%d minutes ago", minutes)
			}
		case duration < 24*time.Hour:
			hours := int(duration.Hours())
			if hours == 1 {
				lastRun = "1 hour ago"
			} else {
				lastRun = fmt.Sprintf("%d hours ago", hours)
			}
		default:
			days := int(duration.Hours() / 24)
			if days == 1 {
				lastRun = "1 day ago"
			} else {
				lastRun = fmt.Sprintf("%d days ago", days)
			}
		}
	}

	return &dto.OptimizationInfo{
		LastRun: lastRun,
		Status:  status,
		RunID: runID,
	}, nil
}

// returns the latest optimizing process for a specific type (MAJOR, MINOR, FULL)
func (s *Service) GetLatestOptimizingProcessByType(ctx context.Context, catalog, database, table, processType string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", constants.OptimisationAPIBase, catalog, database, table)

	params := url.Values{}
	params.Set("type", processType)
	params.Set("page", "1")
	params.Set("pageSize", "1")

	var result map[string]interface{}
	if err := s.DoInto(ctx, http.MethodGet, path, params, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get latest %s process for %s.%s.%s: %s", processType, catalog, database, table, err)
	}

	return result, nil
}
