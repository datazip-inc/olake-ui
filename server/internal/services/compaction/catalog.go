package compaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CatalogRequest represents the request to create or update a catalog
type CatalogRequest struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	OptimizerGroup   string            `json:"optimizerGroup,omitempty"`
	TableFormatList  []string          `json:"tableFormatList"`
	StorageConfig    map[string]string `json:"storageConfig"`
	AuthConfig       map[string]string `json:"authConfig"`
	Properties       map[string]string `json:"properties"`
	TableProperties  map[string]string `json:"tableProperties"`
}

// CatalogResponse represents the response from catalog operations
type CatalogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CreateCatalog creates a new catalog in Amoro
func (c *Compaction) CreateCatalog(ctx context.Context, req CatalogRequest) (*CatalogResponse, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, fmt.Errorf("catalog name is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("catalog type is required")
	}
	if len(req.TableFormatList) == 0 {
		// Default to ICEBERG for Iceberg-specific catalogs
		req.TableFormatList = []string{"ICEBERG"}
	}
	if req.StorageConfig == nil {
		req.StorageConfig = make(map[string]string)
	}
	if req.AuthConfig == nil {
		req.AuthConfig = make(map[string]string)
	}
	if req.Properties == nil {
		req.Properties = make(map[string]string)
	}
	if req.TableProperties == nil {
		req.TableProperties = make(map[string]string)
	}

	// Build the API path
	path := fmt.Sprintf("%scatalogs", apiBase)

	// Make the POST request
	respBody, err := c.doRequest(ctx, http.MethodPost, path, url.Values{}, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog %s: %w", req.Name, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return &CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully created catalog %s", req.Name),
	}, nil
}

// UpdateCatalog updates an existing catalog in Amoro
func (c *Compaction) UpdateCatalog(ctx context.Context, catalogName string, req CatalogRequest) (*CatalogResponse, error) {
	// Validate required fields
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("catalog type is required")
	}
	if len(req.TableFormatList) == 0 {
		// Default to ICEBERG for Iceberg-specific catalogs
		req.TableFormatList = []string{"ICEBERG"}
	}
	if req.StorageConfig == nil {
		req.StorageConfig = make(map[string]string)
	}
	if req.AuthConfig == nil {
		req.AuthConfig = make(map[string]string)
	}
	if req.Properties == nil {
		req.Properties = make(map[string]string)
	}
	if req.TableProperties == nil {
		req.TableProperties = make(map[string]string)
	}

	// Set the name in the request to match the path parameter
	req.Name = catalogName

	// Build the API path
	path := fmt.Sprintf("%scatalogs/%s", apiBase, catalogName)

	// Make the PUT request
	respBody, err := c.doRequest(ctx, http.MethodPut, path, url.Values{}, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update catalog %s: %w", catalogName, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return &CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated catalog %s", catalogName),
	}, nil
}

// DeleteCatalog deletes a catalog from Amoro
func (c *Compaction) DeleteCatalog(ctx context.Context, catalogName string) (*CatalogResponse, error) {
	// Validate required fields
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}

	// Build the API path
	path := fmt.Sprintf("%scatalogs/%s", apiBase, catalogName)

	// Make the DELETE request
	respBody, err := c.doRequest(ctx, http.MethodDelete, path, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to delete catalog %s: %w", catalogName, err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return &CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted catalog %s", catalogName),
	}, nil
}
