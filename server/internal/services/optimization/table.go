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
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

// when no process records exist for the given type.
var errNoProcess = errors.New("no optimizing process found")

// concurrency : bounded by available runtime cores
var tableFanoutWorkers = runtime.NumCPU() * 2

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

	// let each go-routine write to its own index (no mutex required)
	tables := make([]dto.TableInfo, len(tablesList))

	for i, item := range tablesList {
		tableMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid table item type: got %T", item)
		}

		tableName, ok := tableMap["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid table name in %v", tableMap)
		}
		tables[i].Name = tableName

		olakeCreated, ok := tableMap["olake_created"].(bool)
		if !ok {
			return nil, fmt.Errorf("missing or invalid olake created status for table %s.%s.%s", catalog, databaseName, tableName)
		}
		tables[i].OLakeCreated = olakeCreated

		healthScore, ok := tableMap["healthScore"].(float64)
		if !ok {
			return nil, fmt.Errorf("missing or invalid health score for table %s.%s.%s", catalog, databaseName, tableName)
		}
		tables[i].HealthScore = int(healthScore)

		optimizationEnabled, ok := tableMap["enabled"].(bool)
		if !ok {
			return nil, fmt.Errorf("missing or invalid optimization enabled status for table %s.%s.%s", catalog, databaseName, tableName)
		}
		tables[i].Enabled = optimizationEnabled
	}

	// 3 * number of tables : thus processing it concurrently
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(tableFanoutWorkers)
	for i := range tables {
		g.Go(func() error {
			err := s.loadOptimizationInfo(gctx, catalog, databaseName, &tables[i])
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &dto.TablesResponse{
		Catalog:  catalog,
		Database: databaseName,
		Tables:   tables,
	}, nil
}

func (s *Service) loadOptimizationInfo(ctx context.Context, catalog, database string, table *dto.TableInfo) error {
	var res [3]*dto.OptimizationInfo

	processes := [3]string{constants.LiteOptimization, constants.MediumOptimization, constants.FullOptimization}
	g, gctx := errgroup.WithContext(ctx)

	for i, process := range processes {
		g.Go(func() error {
			r, err := s.loadOptimizationInfoForEachTable(gctx, catalog, database, table.Name, utils.ToOptimizationType(process))
			if err != nil && !errors.Is(err, errNoProcess) {
				return fmt.Errorf("failed to fetch latest %s process info: %w", process, err)
			}
			res[i] = r
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	table.Minor = res[0]
	table.Major = res[1]
	table.Full = res[2]

	return nil
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

// loadOptimizationInfoForEachTable fetches the latest optimizing process info for a specific type
func (s *Service) loadOptimizationInfoForEachTable(ctx context.Context, catalog, database, table, processType string) (*dto.OptimizationInfo, error) {
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
