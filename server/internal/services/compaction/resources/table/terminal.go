package table

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

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
func (c *Service) SetTableProperties(ctx context.Context, req models.SetTablePropertiesRequest) (*SetTablePropertiesResponse, error) {
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
	path := fmt.Sprintf("%sterminal/catalogs/%s/execute", models.ApiBase, req.Catalog)

	requestBody := TerminalExecuteRequest{
		SQL: sql,
	}

	respBody, err := c.compaction.DoRequest(ctx, http.MethodPost, path, url.Values{}, requestBody)
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

	// Return immediately with session ID
	return &SetTablePropertiesResponse{
		SessionID: sessionResult.SessionID,
		Status:    "SUBMITTED",
		Success:   true,
		Message:   fmt.Sprintf("ALTER TABLE command submitted. Session ID: %s", sessionResult.SessionID),
	}, nil
}

// EnableSelfOptimizing enables self-optimizing for a table
func (c *Service) EnableSelfOptimizing(ctx context.Context, catalog, database, table string) (*SetTablePropertiesResponse, error) {
	return c.SetTableProperties(ctx, models.SetTablePropertiesRequest{
		Catalog:  catalog,
		Database: database,
		Table:    table,
		Properties: map[string]string{
			"self-optimizing.enabled": "true",
		},
	})
}

// DisableSelfOptimizing disables self-optimizing for a table
func (c *Service) DisableSelfOptimizing(ctx context.Context, catalog, database, table string) (*SetTablePropertiesResponse, error) {
	return c.SetTableProperties(ctx, models.SetTablePropertiesRequest{
		Catalog:  catalog,
		Database: database,
		Table:    table,
		Properties: map[string]string{
			"self-optimizing.enabled": "false",
		},
	})
}
