package compaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CompactionCronConfigRequest represents the request to set compaction cron configuration
type CompactionCronConfigRequest struct {
	Minor          *CronFrequency  `json:"minor,omitempty"`
	Major          *CronFrequency  `json:"major,omitempty"`
	Full           *CronFrequency  `json:"full,omitempty"`
	AdvancedConfig *AdvancedConfig `json:"advanced-config,omitempty"`
	Cleanup        *CleanupConfig  `json:"cleanup,omitempty"`
}

type CronFrequency struct {
	Frequency string `json:"frequency"`
}

type AdvancedConfig struct {
	TargetFileSize string `json:"target-file-size,omitempty"`
}

type CleanupConfig struct {
	AdvancedConfig *CleanupAdvancedConfig `json:"advanced-config,omitempty"`
}

type CleanupAdvancedConfig struct {
	RemoveSnapshotsOlderThan string `json:"remove-snapshots-older-than,omitempty"`
}

// CompactionCronConfigResponse represents the response from setting cron configuration
type CompactionCronConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SetCompactionCronConfig stores the compaction cron configuration in catalog properties
// The configuration is stored with key: table.<database>:<table>:cron
func (c *Compaction) SetCompactionCronConfig(ctx context.Context, catalog, database, table string, config CompactionCronConfigRequest) (*CompactionCronConfigResponse, error) {
	// First, get the current catalog metadata
	catalogPath := fmt.Sprintf("%scatalogs/%s", apiBase, catalog)
	respBody, err := c.doRequest(ctx, http.MethodGet, catalogPath, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog %s: %w", catalog, err)
	}

	var catalogResp Response
	if err := json.Unmarshal(respBody, &catalogResp); err != nil {
		return nil, fmt.Errorf("failed to parse catalog response: %w", err)
	}

	if catalogResp.Code != 200 && catalogResp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", catalogResp.Code, catalogResp.Message)
	}

	// Parse catalog metadata
	var catalogMeta map[string]interface{}
	if err := json.Unmarshal(catalogResp.Result, &catalogMeta); err != nil {
		return nil, fmt.Errorf("failed to parse catalog metadata: %w", err)
	}

	// Get or create properties map
	var properties map[string]interface{}
	if props, ok := catalogMeta["properties"].(map[string]interface{}); ok {
		properties = props
	} else {
		properties = make(map[string]interface{})
		catalogMeta["properties"] = properties
	}

	// Convert config to JSON string
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Store config with key: table.<database>:<table>:cron
	propertyKey := fmt.Sprintf("table.%s:%s:cron", database, table)
	properties[propertyKey] = string(configJSON)

	// Update the catalog with new properties
	updatePath := fmt.Sprintf("%scatalogs/%s", apiBase, catalog)
	updateRespBody, err := c.doRequest(ctx, http.MethodPut, updatePath, url.Values{}, catalogMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to update catalog %s: %w", catalog, err)
	}

	var updateResp Response
	if err := json.Unmarshal(updateRespBody, &updateResp); err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	if updateResp.Code != 200 && updateResp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", updateResp.Code, updateResp.Message)
	}

	return &CompactionCronConfigResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully stored compaction cron configuration for %s.%s.%s", catalog, database, table),
	}, nil
}

// GetCompactionCronConfig retrieves the compaction cron configuration from catalog properties
func (c *Compaction) GetCompactionCronConfig(ctx context.Context, catalog, database, table string) (*CompactionCronConfigRequest, error) {
	// Get the catalog metadata
	catalogPath := fmt.Sprintf("%scatalogs/%s", apiBase, catalog)
	respBody, err := c.doRequest(ctx, http.MethodGet, catalogPath, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog %s: %w", catalog, err)
	}

	var catalogResp Response
	if err := json.Unmarshal(respBody, &catalogResp); err != nil {
		return nil, fmt.Errorf("failed to parse catalog response: %w", err)
	}

	if catalogResp.Code != 200 && catalogResp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", catalogResp.Code, catalogResp.Message)
	}

	// Parse catalog metadata
	var catalogMeta map[string]interface{}
	if err := json.Unmarshal(catalogResp.Result, &catalogMeta); err != nil {
		return nil, fmt.Errorf("failed to parse catalog metadata: %w", err)
	}

	// Get properties map
	properties, ok := catalogMeta["properties"].(map[string]interface{})
	if !ok {
		return &CompactionCronConfigRequest{}, nil
	}

	// Retrieve config with key: table.<database>:<table>:cron
	propertyKey := fmt.Sprintf("table.%s:%s:cron", database, table)
	configStr, ok := properties[propertyKey].(string)
	if !ok || configStr == "" {
		return &CompactionCronConfigRequest{}, nil
	}

	// Parse the JSON config
	var config CompactionCronConfigRequest
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse stored config: %w", err)
	}

	return &config, nil
}
