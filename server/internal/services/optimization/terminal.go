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
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
)

func convertConfigToMap(config dto.OptimizationTableConfig) map[string]string {
	properties := make(map[string]string)
	if config.SQLInput.MinorCron != nil {
		properties[constants.OptMinorCron] = *config.SQLInput.MinorCron
	}
	if config.SQLInput.MajorCron != nil {
		properties[constants.OptMajorCron] = *config.SQLInput.MajorCron
	}
	if config.SQLInput.FullCron != nil {
		properties[constants.OptFullCron] = *config.SQLInput.FullCron
	}
	if config.SQLInput.EnabledForOptimization != nil {
		properties[constants.OptEnableOptimization] = *config.SQLInput.EnabledForOptimization
	}
	if config.SQLInput.TargetFileSize != nil {
		properties[constants.OptTargetFileSize] = utils.ConvertMBToBytes(*config.SQLInput.TargetFileSize)
	}
	return properties
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
func (s *Service) SetProperties(ctx context.Context, catalog, database string, config dto.OptimizationTableConfig) (*dto.TableProperties, error) {
	tables := config.Tables
	properties := convertConfigToMap(config)
	bulkConfigured := len(tables) >= 2

	// Preload metrics + "updated" before ALTER (optional but needed for the contract).
	type tableTelemetry struct {
		name      string
		tableSize string
		fileCount float64
		updated   bool
	}
	pre := make([]tableTelemetry, 0, len(tables))
	for _, tableName := range tables {
		t := tableTelemetry{name: tableName}
		if details, err := s.getTableDetails(ctx, catalog, database, tableName); err == nil {
			if detailsMap, ok := details.(map[string]interface{}); ok {
				if bm, ok := detailsMap["baseMetrics"].(map[string]interface{}); ok {
					if size, ok := bm["totalSize"].(string); ok {
						t.tableSize = size
					}
					if fc, ok := bm["fileCount"].(float64); ok {
						t.fileCount = fc
					}
				}
				if props, ok := detailsMap["properties"].(map[string]interface{}); ok {
					_, hasMinor := props[constants.OptMinorCron]
					_, hasMajor := props[constants.OptMajorCron]
					_, hasFull := props[constants.OptFullCron]
					t.updated = hasMinor || hasMajor || hasFull
				}
			}
		}
		pre = append(pre, t)
	}

	alterTableQuery := make([]string, 0, len(tables))
	for _, tableName := range tables {
		alterTableQuery = append(alterTableQuery, createAlterQuery(database, tableName, properties))
	}

	var sessionResult dto.TerminalSessionResponse
	requestBody := dto.TerminalExecuteRequest{
		SQL: strings.Join(alterTableQuery, "\n"),
	}

	if err := s.DoInto(ctx, http.MethodPost, fmt.Sprintf(constants.OptPathTerminalExecute, catalog), url.Values{}, requestBody, &sessionResult); err != nil {
		return nil, fmt.Errorf("failed to execute bulk ALTER TABLE for catalog %s, database %s: %w", catalog, database, err)
	}

	logInfo, err := s.pollForCompletion(ctx, sessionResult.SessionID)

	if err != nil {
		return nil, fmt.Errorf("failed to poll for completion: %w", err)
	}

	status := "failed"
	if logInfo.LogStatus == "Finished" {
		status = "success"
	}

	// Only emit for configure saves (modal always sends crons + target size).
	if config.SQLInput.MinorCron != nil ||
		config.SQLInput.MajorCron != nil ||
		config.SQLInput.FullCron != nil {

		compactionEnabled := false
		if config.SQLInput.EnabledForOptimization != nil {
			compactionEnabled = *config.SQLInput.EnabledForOptimization == "true"
		}

		for _, t := range pre {
			props := map[string]interface{}{
				"table_size":         t.tableSize,
				"file_count":         t.fileCount,
				"lite_frequency":     properties[constants.OptMinorCron],
				"medium_frequency":   properties[constants.OptMajorCron],
				"full_frequency":     properties[constants.OptFullCron],
				"target_file_size":   properties[constants.OptTargetFileSize],
				"status":             status,
				"bulk_configured":    bulkConfigured,
				"compaction_enabled": compactionEnabled,
				"updated":            t.updated,
			}
			telemetry.TrackConfigurationSaved(ctx, props)
		}
	}

	// TODO: Fusion may return "Finished" even if the query fails, but query logs will contain error message
	return &dto.TableProperties{
		SessionID: sessionResult.SessionID,
		Success:   logInfo.LogStatus == "Finished",
		Logs:      logInfo.Logs,
	}, nil
}

// pollForCompletion polls the terminal API for SQL execution completion
func (s *Service) pollForCompletion(ctx context.Context, sessionID string) (*dto.LogInfo, error) {
	path := fmt.Sprintf(constants.OptPathTerminalLogs, sessionID)
	timeoutCtx, cancel := context.WithTimeout(ctx, constants.OptSessionTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for SQL execution to complete")
		case <-time.After(1 * time.Second):
			var logInfo dto.LogInfo
			if err := s.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &logInfo); err != nil {
				return nil, fmt.Errorf("failed to get logs for session %s: %w", sessionID, err)
			}

			if logInfo.LogStatus == "Running" {
				continue
			}

			return &logInfo, nil
		}
	}
}
