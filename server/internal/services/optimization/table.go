package optimization

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	"golang.org/x/sync/errgroup"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// when no process records exist for the given type.
var errNoProcess = errors.New("no optimizing process found")

// concurrency : bounded by available runtime cores
var tableFanoutWorkers = runtime.NumCPU()

// fetches all tables with full details for a specific catalog and database
func (s *Service) GetTablesWithDetails(ctx context.Context, catalog, databaseName string) (*dto.TablesResponse, error) {
	tablesResult, err := s.getTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, databaseName, err)
	}

	tablesList, ok := tablesResult.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected tables result format for %s.%s: got %T", catalog, databaseName, tablesResult)
	}

	names := make([]string, len(tablesList))
	for i, item := range tablesList {
		tableMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid table item type: got %T", item)
		}
		tableName, ok := tableMap["name"].(string)
		if !ok || tableName == "" {
			return nil, fmt.Errorf("missing or invalid table name in %v", tableMap)
		}
		names[i] = tableName
	}

	// let each go-routine write to its own index (no mutex required)
	results := make([]dto.TableInfo, len(names))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(tableFanoutWorkers)
	for i := range names {
		g.Go(func() error {
			info, err := s.buildTableInfo(gctx, catalog, databaseName, names[i])
			if err != nil {
				return err
			}
			results[i] = info
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &dto.TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   results,
	}, nil
}

func (s *Service) buildTableInfo(ctx context.Context, catalog, database, tableName string) (dto.TableInfo, error) {
	info := dto.TableInfo{
		Name:         tableName,
		Enabled:      false,
		OLakeCreated: false,
	}

	var (
		details            interface{}
		minor, major, full *dto.OptimizationInfo
	)

	g, gctx := errgroup.WithContext(ctx)

	// total : tableFanoutWorkers * 4

	g.Go(func() error {
		d, err := s.getTableDetails(gctx, catalog, database, tableName)
		if err != nil {
			return fmt.Errorf("failed to get details for table %s.%s.%s: %w", catalog, database, tableName, err)
		}
		details = d
		return nil
	})

	processTypes := []struct {
		kind  string
		label string
		dst   **dto.OptimizationInfo
	}{
		{"MINOR", "Lite", &minor},
		{"MAJOR", "Medium", &major},
		{"FULL", "Full", &full},
	}
	for _, pt := range processTypes {
		g.Go(func() error {
			r, err := s.fetchLatestProcessInfo(gctx, catalog, database, tableName, pt.kind)
			if err != nil && !errors.Is(err, errNoProcess) {
				return fmt.Errorf("failed to fetch latest %s process info: %w", pt.label, err)
			}
			*pt.dst = r
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return info, err
	}

	detailsMap, ok := details.(map[string]interface{})
	if !ok {
		return info, fmt.Errorf("invalid tableDetails type: expected map[string]interface{}, got %T", details)
	}

	baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{})
	if !ok {
		return info, fmt.Errorf("missing or invalid baseMetrics for table %s.%s.%s", catalog, database, tableName)
	}

	totalSize, ok := baseMetrics["totalSize"].(string)
	if !ok {
		return info, fmt.Errorf("missing or invalid totalSize in baseMetrics for table %s.%s.%s", catalog, database, tableName)
	}
	info.TotalSize = totalSize

	properties, ok := detailsMap["properties"].(map[string]interface{})
	if !ok {
		return info, fmt.Errorf("missing or invalid properties for table %s.%s.%s", catalog, database, tableName)
	}
	if enabled, ok := properties["self-optimizing.enabled"]; ok {
		if enabledStr, ok := enabled.(string); ok {
			info.Enabled = enabledStr == "true"
		}
	}
	if _, ok := properties["olake_2pc"]; ok {
		info.OLakeCreated = true
	}

	tableSummary, ok := detailsMap["tableSummary"].(map[string]interface{})
	if !ok {
		return info, fmt.Errorf("missing or invalid tableSummary for table %s.%s.%s", catalog, database, tableName)
	}
	healthScore, ok := tableSummary["healthScore"].(float64)
	if !ok {
		return info, fmt.Errorf("missing or invalid healthScore in tableSummary for table %s.%s.%s", catalog, database, tableName)
	}
	info.HealthScore = int(healthScore)

	info.Minor = minor
	info.Major = major
	info.Full = full
	return info, nil
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
