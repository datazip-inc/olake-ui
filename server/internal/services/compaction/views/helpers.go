package aggregator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

func (s *Service) getCatalogMetadata(ctx context.Context, catalog string) (*models.CatalogRequest, error) {
	return s.catalog.GetCatalog(ctx, catalog)
}

// getTableEnabledStatus extracts enabled/disabled status from catalog table properties
// Format: <db>:<tbl> → <enabled>,<minor>,<major>,<full>
func (s *Service) getTableEnabledStatus(catalogMeta *models.CatalogRequest, database, table string) bool {
	if catalogMeta == nil || catalogMeta.TableProperties == nil {
		return false
	}

	tableKey := fmt.Sprintf("%s:%s", database, table)
	configStr, ok := catalogMeta.TableProperties[tableKey]
	if !ok || configStr == "" {
		return false
	}

	// Parse format: <enabled>,<minor>,<major>,<full>
	parts := strings.Split(configStr, ",")
	if len(parts) < 1 {
		return true
	}

	return parts[0] == "true"
}

// fetchLatestProcessInfo fetches the latest optimizing process info for a specific type
func (s *Service) fetchLatestProcessInfo(ctx context.Context, catalog, database, table, processType string) *models.OptimizationInfo {
	result, err := s.table.GetLatestOptimizingProcessByType(ctx, catalog, database, table, processType)
	if err != nil {
		return nil
	}

	processList, ok := result["list"].([]interface{})
	if !ok || len(processList) == 0 {
		return nil
	}

	// Get the first (latest) process
	process, ok := processList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	finishTime, _ := process["finishTime"].(float64)
	status, _ := process["status"].(string)

	var lastRun string
	if finishTime > 0 {
		timestamp := time.Unix(0, int64(finishTime)*int64(time.Millisecond))
		duration := time.Since(timestamp)

		if duration < time.Minute {
			seconds := int(duration.Seconds())
			if seconds == 1 {
				lastRun = "1 sec ago"
			} else {
				lastRun = fmt.Sprintf("%d secs ago", seconds)
			}
		} else if duration < time.Hour {
			minutes := int(duration.Minutes())
			if minutes == 1 {
				lastRun = "1 minute ago"
			} else {
				lastRun = fmt.Sprintf("%d minutes ago", minutes)
			}
		} else {
			hours := int(duration.Hours())
			if hours == 1 {
				lastRun = "1 hour ago"
			} else {
				lastRun = fmt.Sprintf("%d hours ago", hours)
			}
		}
	}

	return &models.OptimizationInfo{
		LastRun: lastRun,
		Status:  status,
	}
}
