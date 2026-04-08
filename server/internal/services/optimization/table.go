package optimization

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// when no process records exist for the given type.
var errNoProcess = errors.New("no optimizing process found")

// fetches all tables with full details for a specific catalog and database
func (s *Service) GetTablesWithDetails(ctx context.Context, catalog, databaseName string) (*dto.TablesResponse, error) {
	response := &dto.TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   make([]dto.TableInfo, 0),
	}

	tablesResult, err := s.getTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, databaseName, err)
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
			Name:         tableName,
			Enabled:      false,
			OLakeCreated: false,
		}

		tableDetails, err := s.getTableDetails(ctx, catalog, databaseName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get details for table %s.%s.%s: %w", catalog, databaseName, tableName, err)
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
			tableInfo.Enabled = enabled.(string) == "true"
		}

		if _, ok := properties["olake_2pc"]; ok {
			tableInfo.OLakeCreated = true
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
		if err != nil && !errors.Is(err, errNoProcess) {
			return nil, fmt.Errorf("failed to fetch latest Lite process info: %w", err)
		}
		tableInfo.Minor = res

		res, err = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "MAJOR")
		if err != nil && !errors.Is(err, errNoProcess) {
			return nil, fmt.Errorf("failed to fetch latest Medium process info: %w", err)
		}
		tableInfo.Major = res

		res, err = s.fetchLatestProcessInfo(ctx, catalog, databaseName, tableName, "FULL")
		if err != nil && !errors.Is(err, errNoProcess) {
			return nil, fmt.Errorf("failed to fetch latest Full process info: %w", err)
		}
		tableInfo.Full = res

		response.Tables = append(response.Tables, tableInfo)
	}

	return response, nil
}

// GetTables returns the list of tables for a given catalog and database
func (s *Service) getTables(ctx context.Context, catalog, database, keywords string) (interface{}, error) {
	path := fmt.Sprintf(constants.OptPathCatalogTables, catalog, database)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	var result interface{}
	err := s.DoInto(ctx, http.MethodGet, path, params, nil, &result)
	return result, err
}

// returns the details of a specific table including size information
func (s *Service) getTableDetails(ctx context.Context, catalog, database, table string) (interface{}, error) {
	path := fmt.Sprintf(constants.OptPathTableDetails, catalog, database, table)

	var result interface{}
	err := s.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &result)
	return result, err
}

// fetchLatestProcessInfo fetches the latest optimizing process info for a specific type
func (s *Service) fetchLatestProcessInfo(ctx context.Context, catalog, database, table, processType string) (*dto.OptimizationInfo, error) {
	result, err := s.getLatestOptimizingProcessByType(ctx, catalog, database, table, processType)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest %s optimizing process for %s.%s.%s: %w", processType, catalog, database, table, err)
	}

	processList, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to fetch optimization process info")
	}

	if len(processList) == 0 {
		return nil, errNoProcess
	}

	// get the first (latest) process
	process, ok := processList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid process format")
	}

	finishTime, _ := process["finishTime"].(float64)
	status, _ := process["status"].(string)
	runID, _ := process["processId"].(string)

	return &dto.OptimizationInfo{
		FinishTime: int64(finishTime),
		Status:     status,
		RunID:      runID,
	}, nil
}

// returns the latest optimizing process for a specific type
func (s *Service) getLatestOptimizingProcessByType(ctx context.Context, catalog, database, table, processType string) (map[string]interface{}, error) {
	path := fmt.Sprintf(constants.OptPathTableOptimizingProcesses, catalog, database, table)

	params := url.Values{}
	params.Set("type", processType)
	params.Set("page", "1")
	params.Set("pageSize", "1")

	var result map[string]interface{}
	if err := s.DoInto(ctx, http.MethodGet, path, params, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get latest %s process for %s.%s.%s: %w", processType, catalog, database, table, err)
	}

	return result, nil
}
