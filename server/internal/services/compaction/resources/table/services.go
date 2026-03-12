package table

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

type Service struct {
	compaction *client.Compaction
}

func NewService(c *client.Compaction) *Service {
	return &Service{
		compaction: c,
	}
}

// GetDatabases returns the list of databases for a given catalog
func (s *Service) GetDatabases(ctx context.Context, catalog, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases", models.APIBase, catalog)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	return s.compaction.Do(ctx, http.MethodGet, path, params, nil)
}

// GetTables returns the list of tables for a given catalog and database
func (s *Service) GetTables(ctx context.Context, catalog, database, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases/%s/tables", models.APIBase, catalog, database)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	return s.compaction.Do(ctx, http.MethodGet, path, params, nil)
}

// returns the details of a specific table including size information
func (s *Service) GetTableDetails(ctx context.Context, catalog, database, table string) (interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/details", models.APIBase, catalog, database, table)

	return s.compaction.Do(ctx, http.MethodGet, path, url.Values{}, nil)
}

// returns the latest snapshot summary for a table
func (s *Service) GetLatestSnapshot(ctx context.Context, catalog, database, table string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/snapshots", models.APIBase, catalog, database, table)

	params := url.Values{}
	params.Set("page", "1")
	params.Set("pageSize", "1")

	result, err := s.compaction.Do(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots for %s.%s.%s: %w", catalog, database, table, err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected snapshots result format for %s.%s.%s", catalog, database, table)
	}

	list, ok := resultMap["list"].([]interface{})
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("no snapshots found for %s.%s.%s", catalog, database, table)
	}

	latestSnapshot, ok := list[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid snapshot format for %s.%s.%s", catalog, database, table)
	}

	summary, ok := latestSnapshot["summary"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no summary found in latest snapshot for %s.%s.%s", catalog, database, table)
	}

	return summary, nil
}

// returns the latest optimizing process for a specific type (MAJOR, MINOR, FULL)
func (s *Service) GetLatestOptimizingProcessByType(ctx context.Context, catalog, database, table, processType string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", models.APIBase, catalog, database, table)

	params := url.Values{}
	params.Set("type", processType)
	params.Set("page", "1")
	params.Set("pageSize", "1")

	var result map[string]interface{}
	if err := s.compaction.DoInto(ctx, http.MethodGet, path, params, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get latest %s process for %s.%s.%s: %w", processType, catalog, database, table, err)
	}

	return result, nil
}
