package optimization

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

func (s *Service) SetCompactionCronConfig(ctx context.Context, catalog, database, table string, config models.CompactionCronConfigRequest) (*models.CompactionCronConfigResponse, error) {
	properties := make(map[string]string)

	minorInterval := parseIntervalValue(config.MinorTriggerInterval)
	if minorInterval != "-1" {
		properties["self-optimizing.minor.trigger.interval"] = minorInterval
	}

	majorInterval := parseIntervalValue(config.MajorTriggerInterval)
	if majorInterval != "-1" {
		// TODO: recheck this
		properties["self-optimizing.minor.trigger.interval"] = majorInterval
	}

	fullInterval := parseIntervalValue(config.FullTriggerInterval)
	if fullInterval != "-1" {
		properties["self-optimizing.full.trigger.interval"] = fullInterval
	}

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

	// store in catalog properties
	if err := s.storeCatalogTableProperty(ctx, catalog, database, table, config); err != nil {
		return nil, fmt.Errorf("failed to store catalog table property: %w", err)
	}

	return &models.CompactionCronConfigResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully configured compaction for %s.%s.%s", catalog, database, table),
	}, nil
}

// retrieves the compaction cron configuration from catalog properties
func (s *Service) GetCompactionCronConfig(ctx context.Context, catalog, database, table string) (*models.CompactionCronConfigRequest, error) {
	path := fmt.Sprintf("%scatalogs/%s", models.APIBase, catalog)

	var catalogMeta map[string]interface{}
	if err := s.compaction.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &catalogMeta); err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}

	tableProperties, _ := catalogMeta["tableProperties"].(map[string]interface{})

	tableKey := fmt.Sprintf("%s:%s", database, table)
	configStr, ok := tableProperties[tableKey].(string)
	if !ok || configStr == "" {
		return &models.CompactionCronConfigRequest{
			MinorTriggerInterval: "never",
			MajorTriggerInterval: "never",
			FullTriggerInterval:  "never",
		}, nil
	}

	// parse the config string: <enabled>,<minor>,<major>,<full>
	parts := strings.Split(configStr, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid config format: expected 4 parts, got %d", len(parts))
	}

	enabled := parts[0] == "true"

	return &models.CompactionCronConfigRequest{
		Enabled:              enabled,
		MinorTriggerInterval: parts[1],
		MajorTriggerInterval: parts[2],
		FullTriggerInterval:  parts[3],
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

// storeCatalogTableProperty stores the configuration in catalog table properties
func (s *Service) storeCatalogTableProperty(ctx context.Context, catalog, database, table string, config models.CompactionCronConfigRequest) error {
	path := fmt.Sprintf("%scatalogs/%s", models.APIBase, catalog)

	var catalogMeta map[string]interface{}
	if err := s.compaction.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &catalogMeta); err != nil {
		return fmt.Errorf("failed to get catalog: %w", err)
	}

	// Get or create table properties map
	var tableProperties map[string]interface{}
	if props, ok := catalogMeta["tableProperties"].(map[string]interface{}); ok {
		tableProperties = props
	} else {
		tableProperties = make(map[string]interface{})
		catalogMeta["tableProperties"] = tableProperties
	}

	// Store config with key: <database>:<table>, value: <enabled>,<minor>,<major>,<full>
	tableKey := fmt.Sprintf("%s:%s", database, table)
	enabledStr := "false"
	if config.Enabled {
		enabledStr = "true"
	}
	configValue := fmt.Sprintf("%s,%s,%s,%s",
		enabledStr,
		parseIntervalValue(config.MinorTriggerInterval),
		parseIntervalValue(config.MajorTriggerInterval),
		parseIntervalValue(config.FullTriggerInterval),
	)
	tableProperties[tableKey] = configValue

	// Update catalog
	updatePath := fmt.Sprintf("%scatalogs/%s", models.APIBase, catalog)
	return s.compaction.DoAndValidate(ctx, http.MethodPut, updatePath, url.Values{}, catalogMeta)
}
