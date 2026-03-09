package table

import (
	"context"
	"encoding/json"
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

type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// GetDatabases returns the list of databases for a given catalog
func (c *Service) GetDatabases(ctx context.Context, catalog string, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases", models.ApiBase, catalog)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	respBody, err := c.compaction.DoRequest(ctx, http.MethodGet, path, params, nil)
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
func (c *Service) GetTables(ctx context.Context, catalog string, database string, keywords string) (interface{}, error) {
	path := fmt.Sprintf("%scatalogs/%s/databases/%s/tables", models.ApiBase, catalog, database)

	params := url.Values{}
	if keywords != "" {
		params.Set("keywords", keywords)
	}

	respBody, err := c.compaction.DoRequest(ctx, http.MethodGet, path, params, nil)
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

// returns the details of a specific table including size information
func (c *Service) GetTableDetails(ctx context.Context, catalog string, database string, table string) (interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/details", models.ApiBase, catalog, database, table)

	return c.compaction.Do(ctx, http.MethodGet, path, url.Values{}, nil)
}

// returns the latest snapshot summary for a table
func (c *Service) GetLatestSnapshot(ctx context.Context, catalog string, database string, table string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/snapshots", models.ApiBase, catalog, database, table)

	params := url.Values{}
	params.Set("page", "1")
	params.Set("pageSize", "1")

	result, err := c.compaction.Do(ctx, http.MethodGet, path, params, nil)
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

// GetOptimizingProcesses returns the optimization process history for a table
func (c *Service) GetOptimizingProcesses(ctx context.Context, catalog string, database string, table string) (interface{}, error) {
	path := fmt.Sprintf("%stables/catalogs/%s/dbs/%s/tables/%s/optimizing-processes", models.ApiBase, catalog, database, table)

	params := url.Values{}

	respBody, err := c.compaction.DoRequest(ctx, http.MethodGet, path, params, nil)
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
