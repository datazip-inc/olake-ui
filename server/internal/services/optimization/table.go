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
var tableFanoutWorkers = runtime.NumCPU() * 2

// amoroTableMeta mirrors a single entry of Amoro's table-list payload.
// Pointer fields let us distinguish "absent" (legacy Amoro, name/type only)
// from a real zero value present in the enriched (new) payload.
type amoroTableMeta struct {
	Name         string           `json:"name"`
	Type         string           `json:"type"`
	HealthScore  *int             `json:"healthScore"`
	Enabled      *bool            `json:"enabled"`
	OLakeCreated *bool            `json:"olake_created"`
	Lite         *amoroCompaction `json:"lite"`
	Medium       *amoroCompaction `json:"medium"`
	Full         *amoroCompaction `json:"full"`
}

type amoroCompaction struct {
	RunID      string `json:"run_id"`
	FinishTime int64  `json:"finish_time"`
	Status     string `json:"status"`
}

func (c *amoroCompaction) toOptimizationInfo() *dto.OptimizationInfo {
	if c == nil {
		return nil
	}
	return &dto.OptimizationInfo{
		FinishTime: c.FinishTime,
		Status:     c.Status,
		RunID:      c.RunID,
	}
}

func (m amoroTableMeta) toTableInfo() dto.TableInfo {
	// totalSize is intentionally omitted: the enriched API does not return it.
	info := dto.TableInfo{
		Name:  m.Name,
		Minor: m.Lite.toOptimizationInfo(),
		Major: m.Medium.toOptimizationInfo(),
		Full:  m.Full.toOptimizationInfo(),
	}
	if m.HealthScore != nil {
		info.HealthScore = *m.HealthScore
	}
	if m.Enabled != nil {
		info.Enabled = *m.Enabled
	}
	if m.OLakeCreated != nil {
		info.OLakeCreated = *m.OLakeCreated
	}
	return info
}

// GetTablesWithDetails returns all tables for a catalog/database with optimizing
// details. Newer Amoro builds enrich the table-list payload with the per-table
// optimizing metadata we need, so we consume it directly (single upstream call).
// Older builds only return name/type, in which case we fall back to per-table
// detail fan-out. Capability is detected from the response shape so no version
// coordination is required and rolling upgrades are handled transparently.
func (s *Service) GetTablesWithDetails(ctx context.Context, catalog, databaseName string) (*dto.TablesResponse, error) {
	tables, err := s.listTables(ctx, catalog, databaseName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for catalog %s, database %s: %w", catalog, databaseName, err)
	}

	// Only the enriched (new) payload populates "enabled"; its absence means
	// the upstream Amoro is a legacy build and we must fan out for details.
	if len(tables) > 0 && tables[0].Enabled != nil {
		results := make([]dto.TableInfo, len(tables))
		for i := range tables {
			results[i] = tables[i].toTableInfo()
		}
		return &dto.TablesResponse{
			Catalog:  catalog,
			Database: databaseName,
			Tables:   results,
		}, nil
	}

	return s.buildTablesWithDetails(ctx, catalog, databaseName, tables)
}

// buildTablesWithDetails fans out per-table detail lookups (legacy Amoro path).
func (s *Service) buildTablesWithDetails(ctx context.Context, catalog, databaseName string, tables []amoroTableMeta) (*dto.TablesResponse, error) {
	// let each go-routine write to its own index (no mutex required)
	results := make([]dto.TableInfo, len(tables))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(tableFanoutWorkers)
	for i := range tables {
		name := tables[i].Name
		if name == "" {
			return nil, fmt.Errorf("missing or invalid table name in %+v", tables[i])
		}
		g.Go(func() error {
			info, err := s.buildTableInfo(gctx, catalog, databaseName, name)
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
		details interface{}
		procs   [3]*dto.OptimizationInfo
	)
	processKinds := [3]string{"MINOR", "MAJOR", "FULL"}
	processLabels := [3]string{"Lite", "Medium", "Full"}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		d, err := s.getTableDetails(gctx, catalog, database, tableName)
		if err != nil {
			return fmt.Errorf("failed to get details for table %s.%s.%s: %w", catalog, database, tableName, err)
		}
		details = d
		return nil
	})

	for i, kind := range processKinds {
		g.Go(func() error {
			r, err := s.fetchLatestProcessInfo(gctx, catalog, database, tableName, kind)
			if err != nil && !errors.Is(err, errNoProcess) {
				return fmt.Errorf("failed to fetch latest %s process info: %w", processLabels[i], err)
			}
			procs[i] = r
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

	info.Minor = procs[0]
	info.Major = procs[1]
	info.Full = procs[2]
	return info, nil
}

// listTables returns the table-list payload for a catalog/database, decoded into
// amoroTableMeta. Enriched fields are only populated by newer Amoro builds.
func (s *Service) listTables(ctx context.Context, catalog, database, keywords string) ([]amoroTableMeta, error) {
	path := fmt.Sprintf(constants.OptPathCatalogTables, catalog, database)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	var result []amoroTableMeta
	if err := s.DoInto(ctx, http.MethodGet, path, params, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
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
