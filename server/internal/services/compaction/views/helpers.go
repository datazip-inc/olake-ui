package aggregator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

// getTableEnabledStatus extracts enabled/disabled status from catalog table properties
// Format: <db>:<tbl> → <enabled>,<minor>,<major>,<full>
func (s *Service) getTableEnabledStatus(catalogMeta *models.CatalogRequest, database, table string) bool {
	if catalogMeta.TableProperties == nil {
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
		return false
	}

	return parts[0] == "true"
}

// fetchLatestProcessInfo fetches the latest optimizing process info for a specific type
func (s *Service) fetchLatestProcessInfo(ctx context.Context, catalog, database, table, processType string) (*models.OptimizationInfo, error) {
	result, err := s.table.GetLatestOptimizingProcessByType(ctx, catalog, database, table, processType)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	processList, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to fetch compaction process info")
	}

	// Return nil if no processes exist
	if len(processList) == 0 {
		return nil, nil
	}

	// Get the first (latest) process
	process, ok := processList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid process format")
	}

	finishTime, _ := process["finishTime"].(float64)
	status, _ := process["status"].(string)

	var lastRun string
	if finishTime > 0 {
		timestamp := time.Unix(0, int64(finishTime)*int64(time.Millisecond))
		duration := time.Since(timestamp)

		switch {
		case duration < time.Minute:
			seconds := int(duration.Seconds())
			if seconds == 1 {
				lastRun = "1 sec ago"
			} else {
				lastRun = fmt.Sprintf("%d secs ago", seconds)
			}
		case duration < time.Hour:
			minutes := int(duration.Minutes())
			if minutes == 1 {
				lastRun = "1 minute ago"
			} else {
				lastRun = fmt.Sprintf("%d minutes ago", minutes)
			}
		default:
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
	}, nil
}
