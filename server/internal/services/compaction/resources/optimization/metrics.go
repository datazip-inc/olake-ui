package optimization

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/table"
)

type Service struct {
	compaction *client.Compaction
	table      *table.Service
}

func NewService(c *client.Compaction, tbl *table.Service) *Service {
	return &Service{
		compaction: c,
		table:      tbl,
	}
}

// TableMetricsResponse represents detailed metrics for a table
type TableMetricsResponse struct {
	TableMetrics TableMetrics `json:"table-metrics"`
}

type TableMetrics struct {
	FileCount       FileCount `json:"file-count"`
	AverageFileSize string    `json:"average-file-size"`
	LastCommitTime  int64     `json:"last-commit-time,omitempty"`
}

type FileCount struct {
	Total       int `json:"total"`
	DataFiles   int `json:"data-files"`
	DeleteFiles int `json:"delete-files"`
}

type DeleteFiles struct {
	Equality   int `json:"equality"`
	Positional int `json:"positional"`
}

// fetches detailed file metrics for a specific table.
func (s *Service) GetTableMetrics(ctx context.Context, catalog, database, table string) (*TableMetricsResponse, error) {
	tableDetails, err := s.table.GetTableDetails(ctx, catalog, database, table)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get table details for %s.%s.%s: %w",
			catalog, database, table, err,
		)
	}

	response := &TableMetricsResponse{}

	detailsMap, ok := tableDetails.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected table details format")
	}

	baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing baseMetrics in table details")
	}

	if fileCount, ok := baseMetrics["fileCount"].(float64); ok {
		response.TableMetrics.FileCount.Total = int(fileCount)
	}

	if avgSize, ok := baseMetrics["averageFileSize"].(string); ok {
		response.TableMetrics.AverageFileSize = avgSize
	}

	if lastCommitTime, ok := baseMetrics["lastCommitTime"].(float64); ok {
		response.TableMetrics.LastCommitTime = int64(lastCommitTime)
	}

	// from the latest snapshot
	snapshotSummary, err := s.table.GetLatestSnapshot(ctx, catalog, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest snapshot for %s.%s.%s: %w", catalog, database, table, err)
	}

	if snapshotSummary != nil {
		if dataFilesStr, ok := snapshotSummary["total-data-files"].(string); ok {
			if count, err := parseIntFromString(dataFilesStr); err == nil {
				response.TableMetrics.FileCount.DataFiles = count
			}
		}

		if deleteFilesStr, ok := snapshotSummary["total-delete-files"].(string); ok {
			if count, err := parseIntFromString(deleteFilesStr); err == nil {
				response.TableMetrics.FileCount.DeleteFiles = count
			}
		}
	}

	return response, nil
}

// helper function
func parseIntFromString(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
