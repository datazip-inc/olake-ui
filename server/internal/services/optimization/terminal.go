package optimization

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

func setPropertiesMap(config dto.SQLInput) map[string]string {
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
		properties[constants.OptTargetFileSize] = utils.ConvertMBToBytes(*config.TargetFileSize)
	}
	return properties
}

func (s *Service) SetProperties(ctx context.Context, catalog, database, table string, config dto.SQLInput) (*dto.TableProperties, error) {
	properties := setPropertiesMap(config)
	sql := createAlterQuery(database, table, properties)

	path := fmt.Sprintf(constants.OptPathTerminalExecute, catalog)
	requestBody := dto.TerminalExecuteRequest{
		SQL: sql,
	}

	var sessionResult dto.TerminalSessionResponse
	if err := s.DoInto(ctx, http.MethodPost, path, url.Values{}, requestBody, &sessionResult); err != nil {
		return nil, fmt.Errorf("failed to execute ALTER TABLE for %s.%s.%s: %w", catalog, database, table, err)
	}
	// TODO: Fusion may return "Finished" even if the query fails (e.g., syntax error).
	// Solution: validate execution status by checking logs for "Finished" vs "Failed".
	return s.pollAndBuild(ctx, sessionResult.SessionID, catalog, database)
}

func createAlterQuery(database, table string, properties map[string]string) string {
	props := make([]string, 0, len(properties))
	for k, value := range properties {
		props = append(props, fmt.Sprintf("'%s'='%s'", k, value))
	}
	propsJoined := strings.Join(props, ", ")

	return fmt.Sprintf(constants.OptSQLCommand, database, table, propsJoined) + ";"
}

// set properties for multiple tables using sql query
func (s *Service) BulkSetProperties(ctx context.Context, catalog, database string, req dto.BulkSQLInput) (*dto.TableProperties, error) {
	tables := req.Tables
	properties := setPropertiesMap(req.SQLInput)
	// Bulk apply always enables self-optimizing on every selected table.
	properties[constants.OptEnableOptimization] = "true"

	sql := make([]string, 0, len(tables))
	for _, t := range tables {
		sql = append(sql, createAlterQuery(database, t, properties))
	}
	bulkSQLScript := strings.Join(sql, "\n")

	path := fmt.Sprintf(constants.OptPathTerminalExecute, catalog)
	requestBody := dto.TerminalExecuteRequest{
		SQL: bulkSQLScript,
	}

	var sessionResult dto.TerminalSessionResponse
	if err := s.DoInto(ctx, http.MethodPost, path, url.Values{}, requestBody, &sessionResult); err != nil {
		return nil, fmt.Errorf("failed to execute bulk ALTER TABLE for catalog %s, database %s: %w", catalog, database, err)
	}

	return s.pollAndBuild(ctx, sessionResult.SessionID, catalog, database)
}

//  polls for SQL execution completion and constructs a TableProperties response.
func (s *Service) pollAndBuild(ctx context.Context, sessionID, catalog, database string) (*dto.TableProperties, error) {
	logInfo, err := s.pollForCompletion(ctx, sessionID)
	
	if err != nil {
		return nil, fmt.Errorf("failed to poll for completion: %w", err)
	}
	success := logInfo.LogStatus == "Finished"
	var message string
	if success {
		message = fmt.Sprintf(
			"optimization sql command completed successfully for catalog %s, database %s, session ID: %s",
			catalog,
			database,
			sessionID,
		)
	} else {
		message = fmt.Sprintf(
			"optimization sql command failed for catalog %s, database %s, session ID: %s",
			catalog,
			database,
			sessionID,
		)
	}
	return &dto.TableProperties{
		SessionID: sessionID,
		Success:   success,
		Message:   message,
		Logs:      logInfo.Logs,
	}, nil
}

// pollForCompletion polls the terminal API for SQL execution completion
func (s *Service) pollForCompletion(ctx context.Context, sessionID string) (*dto.LogInfo, error) {
	path := fmt.Sprintf(constants.OptPathTerminalLogs, sessionID)
	timeoutCtx, cancel := context.WithTimeout(ctx, constants.OptSessionTimeout)
	defer cancel()

	ticker := time.NewTicker(constants.OptQueryResultPollTime)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for SQL execution to complete")
		case <-ticker.C:
			var logInfo dto.LogInfo
			if err := s.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &logInfo); err != nil {
				return nil, fmt.Errorf("failed to get logs for session %s: %w", sessionID, err)
			}
			return &logInfo, nil
		}
	}
}
