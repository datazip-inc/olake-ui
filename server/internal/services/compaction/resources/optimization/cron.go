package optimization

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

func (s *Service) SetCompactionCronConfig(ctx context.Context, catalog, database, table string, config models.CompactionCronConfigRequest) (*models.CompactionCronConfigResponse, error) {
	properties := make(map[string]string)

	minorInterval := parseIntervalValue(config.MinorTriggerInterval)
	properties["self-optimizing.minor.trigger.interval"] = minorInterval

	majorInterval := parseIntervalValue(config.MajorTriggerInterval)
	properties["self-optimizing.major.trigger.interval"] = majorInterval

	fullInterval := parseIntervalValue(config.FullTriggerInterval)
	properties["self-optimizing.full.trigger.interval"] = fullInterval

	// sql query
	sqlResult, err := s.table.SetTableProperties(ctx, models.SetTablePropertiesRequest{
		Catalog:    catalog,
		Database:   database,
		Table:      table,
		Properties: properties,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL for table properties: %w", err)
	}

	if !sqlResult.Success {
		return &models.CompactionCronConfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to set table properties via SQL: %s", sqlResult.Message),
		}, nil
	}

	return &models.CompactionCronConfigResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully configured compaction for %s.%s.%s", catalog, database, table),
	}, nil
}

// retrieves the compaction cron configuration from table details properties
func (s *Service) GetCompactionCronConfig(ctx context.Context, catalog, database, table string) (*models.CompactionCronConfigRequest, error) {
	tableDetails, err := s.table.GetTableDetails(ctx, catalog, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table details for %s.%s.%s: %w", catalog, database, table, err)
	}

	detailsMap, ok := tableDetails.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tableDetails type: expected map[string]interface{}, got %T", tableDetails)
	}

	properties, ok := detailsMap["properties"].(map[string]interface{})
	if !ok {
		return &models.CompactionCronConfigRequest{
			MinorTriggerInterval: "-1",
			MajorTriggerInterval: "-1",
			FullTriggerInterval:  "-1",
		}, nil
	}

	enabled := false
	if enabledVal, ok := properties["self-optimizing.enabled"]; ok {
		enabled = enabledVal.(string) == "true"
	}

	if !enabled {
		return &models.CompactionCronConfigRequest{
			MinorTriggerInterval: "-1",
			MajorTriggerInterval: "-1",
			FullTriggerInterval:  "-1",
		}, nil
	}

	minorInterval := "-1"
	if val, ok := properties["self-optimizing.minor.trigger.interval"]; ok {
		if strVal, ok := val.(string); ok {
			minorInterval = strVal
		}
	}

	majorInterval := "-1"
	if val, ok := properties["self-optimizing.major.trigger.interval"]; ok {
		if strVal, ok := val.(string); ok {
			majorInterval = strVal
		}
	}

	fullInterval := "-1"
	if val, ok := properties["self-optimizing.full.trigger.interval"]; ok {
		if strVal, ok := val.(string); ok {
			fullInterval = strVal
		}
	}

	return &models.CompactionCronConfigRequest{
		MinorTriggerInterval: minorInterval,
		MajorTriggerInterval: majorInterval,
		FullTriggerInterval:  fullInterval,
	}, nil
}

// parseIntervalValue converts user input to milliseconds or -1 for "never"
func parseIntervalValue(interval string) string {
	// never
	if strings.EqualFold(interval, "never") || interval == "" {
		return "-1"
	}

	if _, err := strconv.ParseInt(interval, 10, 64); err == nil {
		return interval
	}

	return "-1"
}
