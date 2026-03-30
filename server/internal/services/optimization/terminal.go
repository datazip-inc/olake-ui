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

func (s *Service) SetProperties(ctx context.Context, catalog, database, table string, config dto.SQLInput) (*dto.TableProperties, error) {
	properties := make(map[string]string)

	if config.MinorCron != nil {
		properties[constants.OptMinorCron] = *config.MinorCron
	}
	if config.MajorCron != nil {
		properties[constants.OptMajorCron] = *config.MajorCron
	}
	if config.FullCron != nil {
		properties[constants.OptFullCron] = *config.FullCron
	}
	if config.EnabledForOptimization != nil {
		properties[constants.OptEnableOptimization] = *config.EnabledForOptimization
	}
	if config.TargetFileSize != nil {
		properties[constants.OptTargetFileSize] = ConvertMBToBytes(*config.TargetFileSize)
	}

	// sql query
	sqlResult, err := s.SetTableProperties(ctx, dto.SetTablePropertiesRequest{
		Catalog:    catalog,
		Database:   database,
		Table:      table,
		Properties: properties,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to set optimization properties: %s", err)
	}

	return sqlResult, nil
}

// sets table properties using the SQL query
func (s *Service) SetTableProperties(ctx context.Context, req dto.SetTablePropertiesRequest) (*dto.TableProperties, error) {
	var propsSQL []string
	for key, value := range req.Properties {
		propsSQL = append(propsSQL, fmt.Sprintf("'%s' = '%s'", key, value))
	}

	sql := fmt.Sprintf(constants.OptSQLCommand, req.Database, req.Table, strings.Join(propsSQL, ", "))

	// execute via Terminal API
	path := fmt.Sprintf(constants.OptPathTerminalExecute, req.Catalog)
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

	// TODO: Fusion may return "Finished" even if the query fails (e.g., syntax error).
	// Solution: validate execution status by checking logs for "Finished" vs "Failed".
	success := logInfo.LogStatus == "Finished"
	var message string
	if success {
		message = fmt.Sprintf("optimization sql command completed successfully. Session ID: %s", sessionResult.SessionID)
	} else {
		message = fmt.Sprintf("optimization sql command failed with status: %s. Session ID: %s", logInfo.LogStatus, sessionResult.SessionID)
	}

	return &dto.TableProperties{
		SessionID: sessionResult.SessionID,
		Status:    logInfo.LogStatus,
		Success:   success,
		Message:   message,
		Logs:      logInfo.Logs,
	}, nil
}

// pollForCompletion polls the terminal API for SQL execution completion
func (s *Service) pollForCompletion(ctx context.Context, _, sessionID string) (*dto.LogInfo, error) {
	path := fmt.Sprintf(constants.OptPathTerminalLogs, sessionID)
	timeoutCtx, cancel := context.WithTimeout(ctx, constants.OptMaxTimeout)
	defer cancel()

	ticker := time.NewTicker(constants.PollInterval)
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
