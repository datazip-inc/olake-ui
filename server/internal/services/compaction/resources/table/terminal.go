package table

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

// SetTablePropertiesResponse represents the response from setting table properties
type SetTablePropertiesResponse struct {
	SessionID string   `json:"sessionId"`
	Status    string   `json:"status"`
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Logs      []string `json:"logs,omitempty"`
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

// LogInfo represents the log information from terminal execution
type LogInfo struct {
	LogStatus string   `json:"logStatus"`
	Logs      []string `json:"logs"`
}

// sets table properties using the SQL query
func (s *Service) SetTableProperties(ctx context.Context, req models.SetTablePropertiesRequest) (*SetTablePropertiesResponse, error) {
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
	path := fmt.Sprintf("%sterminal/catalogs/%s/execute", models.APIBase, req.Catalog)
	requestBody := TerminalExecuteRequest{
		SQL: sql,
	}

	var sessionResult TerminalSessionResponse
	if err := s.compaction.DoInto(ctx, http.MethodPost, path, url.Values{}, requestBody, &sessionResult); err != nil {
		return nil, fmt.Errorf("failed to execute ALTER TABLE for %s.%s.%s: %w", req.Catalog, req.Database, req.Table, err)
	}

	// Poll for execution completion
	logInfo, err := s.pollForCompletion(ctx, req.Catalog, sessionResult.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for completion: %w", err)
	}

	// Determine success based on status
	success := logInfo.LogStatus == "Finished"
	var message string
	if success {
		message = fmt.Sprintf("ALTER TABLE command completed successfully. Session ID: %s", sessionResult.SessionID)
	} else {
		message = fmt.Sprintf("ALTER TABLE command failed with status: %s. Session ID: %s", logInfo.LogStatus, sessionResult.SessionID)
	}

	return &SetTablePropertiesResponse{
		SessionID: sessionResult.SessionID,
		Status:    logInfo.LogStatus,
		Success:   success,
		Message:   message,
		Logs:      logInfo.Logs,
	}, nil
}

func (s *Service) EnableSelfOptimizing(ctx context.Context, catalog, database, table string) (*SetTablePropertiesResponse, error) {
	return s.SetTableProperties(ctx, models.SetTablePropertiesRequest{
		Catalog:  catalog,
		Database: database,
		Table:    table,
		Properties: map[string]string{
			"self-optimizing.enabled": "true",
		},
	})
}

func (s *Service) DisableSelfOptimizing(ctx context.Context, catalog, database, table string) (*SetTablePropertiesResponse, error) {
	return s.SetTableProperties(ctx, models.SetTablePropertiesRequest{
		Catalog:  catalog,
		Database: database,
		Table:    table,
		Properties: map[string]string{
			"self-optimizing.enabled": "false",
		},
	})
}

// pollForCompletion polls the Amoro server for SQL execution completion
func (s *Service) pollForCompletion(ctx context.Context, _, sessionID string) (*LogInfo, error) {
	const (
		pollInterval = 1500 * time.Millisecond
		maxTimeout   = 5 * time.Minute
	)

	path := fmt.Sprintf("%sterminal/%s/logs", models.APIBase, sessionID)
	timeoutCtx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for SQL execution to complete")
		case <-ticker.C:
			var logInfo LogInfo
			if err := s.compaction.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &logInfo); err != nil {
				return nil, fmt.Errorf("failed to get logs for session %s: %w", sessionID, err)
			}

			// Check if execution is complete
			if logInfo.LogStatus == "Finished" || logInfo.LogStatus == "Failed" ||
				logInfo.LogStatus == "Canceled" || logInfo.LogStatus == "Expired" {
				return &logInfo, nil
			}
		}
	}
}
