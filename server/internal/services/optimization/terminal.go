package optimization

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

func (s *Service) SetProperties(ctx context.Context, catalog, database, table string, config dto.SQLInput) (*dto.SetTablePropertiesResponse, error) {
	properties := make(map[string]string)

	if config.MinorTriggerInterval != "" {
		properties["self-optimizing.minor.trigger.interval"] = config.MinorTriggerInterval
	}

	if config.MajorTriggerInterval != "" {
		properties["self-optimizing.major.trigger.interval"] = config.MajorTriggerInterval
	}

	if config.FullTriggerInterval != "" {
		properties["self-optimizing.full.trigger.interval"] = config.FullTriggerInterval
	}

	if config.TargetFileSize > 0 {
		size := config.TargetFileSize
		targetFileSize := ConvertMBToBytes(int64(size))
		properties["write.target-file-size-bytes"] = targetFileSize
	}

	if config.EnabledForOptimization != "" {
		properties["self-optimizing.enabled"] = config.EnabledForOptimization
	}

	// sql query
	sqlResult, err := s.SetTableProperties(ctx, dto.SetTablePropertiesRequest{
		Catalog:    catalog,
		Database:   database,
		Table:      table,
		Properties: properties,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL for table properties: %s", err)
	}

	return sqlResult, nil
}

// sets table properties using the SQL query
func (s *Service) SetTableProperties(ctx context.Context, req dto.SetTablePropertiesRequest) (*dto.SetTablePropertiesResponse, error) {
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
	path := fmt.Sprintf("%sterminal/catalogs/%s/execute", constants.OptimizationAPIBase, req.Catalog)
	requestBody := dto.TerminalExecuteRequest{
		SQL: sql,
	}

	var sessionResult dto.TerminalSessionResponse
	if err := s.DoInto(ctx, http.MethodPost, path, url.Values{}, requestBody, &sessionResult); err != nil {
		return nil, fmt.Errorf("failed to execute ALTER TABLE for %s.%s.%s: %s", req.Catalog, req.Database, req.Table, err)
	}

	// Poll for execution completion
	logInfo, err := s.pollForCompletion(ctx, req.Catalog, sessionResult.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for completion: %s", err)
	}

	// Determine success based on status
	success := logInfo.LogStatus == "Finished"
	var message string
	if success {
		message = fmt.Sprintf("ALTER TABLE command completed successfully. Session ID: %s", sessionResult.SessionID)
	} else {
		message = fmt.Sprintf("ALTER TABLE command failed with status: %s. Session ID: %s", logInfo.LogStatus, sessionResult.SessionID)
	}

	return &dto.SetTablePropertiesResponse{
		SessionID: sessionResult.SessionID,
		Status:    logInfo.LogStatus,
		Success:   success,
		Message:   message,
		Logs:      logInfo.Logs,
	}, nil
}

// pollForCompletion polls the terminal API for SQL execution completion
func (s *Service) pollForCompletion(ctx context.Context, _, sessionID string) (*dto.LogInfo, error) {
	const (
		pollInterval = 1500 * time.Millisecond
		maxTimeout   = 30 * time.Second
	)

	path := fmt.Sprintf("%sterminal/%s/logs", constants.OptimizationAPIBase, sessionID)
	timeoutCtx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for SQL execution to complete")
		case <-ticker.C:
			var logInfo dto.LogInfo
			if err := s.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &logInfo); err != nil {
				return nil, fmt.Errorf("failed to get logs for session %s: %s", sessionID, err)
			}

			// Check if execution is complete
			if logInfo.LogStatus == "Finished" || logInfo.LogStatus == "Failed" ||
				logInfo.LogStatus == "Canceled" || logInfo.LogStatus == "Expired" {
				return &logInfo, nil
			}
		}
	}
}

func ConvertMBToBytes(sizeMB int64) string {
	const bytesPerMB = 1024 * 1024
	sizeBytes := sizeMB * bytesPerMB
	return strconv.FormatInt(sizeBytes, 10)
}
