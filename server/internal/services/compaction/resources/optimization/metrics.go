package optimization

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	tbl "github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/tables"
)

type Service struct {
	compaction *client.Compaction
}

func NewService(c *client.Compaction) *Service {
	return &Service{
		compaction: c,
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
	Total       int         `json:"total"`
	DataFiles   int         `json:"data-files"`
	DeleteFiles DeleteFiles `json:"delete-files"`
}

type DeleteFiles struct {
	Equality   int `json:"equality"`
	Positional int `json:"positional"`
}

// GetTableMetrics fetches detailed file metrics for a specific table
func (c *Service) GetTableMetrics(ctx context.Context, catalog, database, table string) (*TableMetricsResponse, error) {
	t := tbl.NewService(c.compaction)
	tableDetails, err := t.GetTableDetails(ctx, catalog, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table details for %s.%s.%s: %w", catalog, database, table, err)
	}

	response := &TableMetricsResponse{
		TableMetrics: TableMetrics{
			FileCount: FileCount{
				DeleteFiles: DeleteFiles{
					Equality:   0,
					Positional: 0,
				},
			},
		},
	}

	if detailsMap, ok := tableDetails.(map[string]interface{}); ok {
		if baseMetrics, ok := detailsMap["baseMetrics"].(map[string]interface{}); ok {
			if fileCount, ok := baseMetrics["fileCount"].(float64); ok {
				response.TableMetrics.FileCount.Total = int(fileCount)
			}

			if avgSize, ok := baseMetrics["averageFileSize"].(string); ok {
				response.TableMetrics.AverageFileSize = avgSize
			}

			if lastCommitTime, ok := baseMetrics["lastCommitTime"].(float64); ok {
				response.TableMetrics.LastCommitTime = int64(lastCommitTime)
			}
		}

		if tableSummary, ok := detailsMap["tableSummary"].(map[string]interface{}); ok {
			if summary, ok := tableSummary["summary"].(map[string]interface{}); ok {
				if dataFiles, ok := summary["total-data-files"].(string); ok {
					if count, err := parseIntFromString(dataFiles); err == nil {
						response.TableMetrics.FileCount.DataFiles = count
					}
				} else if dataFiles, ok := summary["total-data-files"].(float64); ok {
					response.TableMetrics.FileCount.DataFiles = int(dataFiles)
				}

				if deleteFiles, ok := summary["total-delete-files"].(string); ok {
					if count, err := parseIntFromString(deleteFiles); err == nil {
						totalDeleteFiles := count

						if eqDeletes, ok := summary["total-equality-deletes"].(string); ok {
							if eqCount, err := parseIntFromString(eqDeletes); err == nil {
								response.TableMetrics.FileCount.DeleteFiles.Equality = eqCount
							}
						}

						if posDeletes, ok := summary["total-positional-deletes"].(string); ok {
							if posCount, err := parseIntFromString(posDeletes); err == nil {
								response.TableMetrics.FileCount.DeleteFiles.Positional = posCount
							}
						}

						if response.TableMetrics.FileCount.DeleteFiles.Equality == 0 &&
							response.TableMetrics.FileCount.DeleteFiles.Positional == 0 {
							response.TableMetrics.FileCount.DeleteFiles.Equality = totalDeleteFiles
						}
					}
				} else if deleteFiles, ok := summary["total-delete-files"].(float64); ok {
					response.TableMetrics.FileCount.DeleteFiles.Equality = int(deleteFiles)
				}
			}
		}

		// If we couldn't get data files from summary, calculate from total - delete files
		if response.TableMetrics.FileCount.DataFiles == 0 && response.TableMetrics.FileCount.Total > 0 {
			totalDeleteFiles := response.TableMetrics.FileCount.DeleteFiles.Equality +
				response.TableMetrics.FileCount.DeleteFiles.Positional
			response.TableMetrics.FileCount.DataFiles = response.TableMetrics.FileCount.Total - totalDeleteFiles
		}
	}

	return response, nil
}

// Helper function to parse integer from string
func parseIntFromString(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
