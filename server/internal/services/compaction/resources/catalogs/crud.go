package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
)

type Service struct {
	compaction *client.Compaction
}

func NewService(c *client.Compaction) *Service {
	return &Service{
		compaction: c,
	}
}

// GetCatalogs returns the list of catalogs from fusion
func (c *Service) GetCatalogs(ctx context.Context) (interface{}, error) {
	path := models.ApiBase + "catalogs"
	respBody, err := c.compaction.DoRequest(ctx, http.MethodGet, path, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all catalogs: %w", err)
	}

	var resp models.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	// Result can be an array or object, so unmarshal to interface{}
	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse catalog result: %w", err)
	}

	return result, nil
}

// CreateCatalog creates a new catalog in Fusion, need to be very particular here
func (c *Service) CreateCatalog(ctx context.Context, req models.CatalogRequest) (*models.CatalogResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("catalog name is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("catalog type is required")
	}

	req.TableFormatList = "ICEBERG"
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

	path := fmt.Sprintf("%scatalogs", models.ApiBase)

	respBody, err := c.compaction.DoRequest(ctx, http.MethodPost, path, url.Values{}, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog %s: %w", req.Name, err)
	}

	var resp models.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully created catalog %s", req.Name),
	}, nil
}

// UpdateCatalog updates an existing catalog in Amoro
func (c *Service) UpdateCatalog(ctx context.Context, catalogName string, req models.CatalogRequest) (*models.CatalogResponse, error) {
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("catalog type is required")
	}

	req.TableFormatList = "ICEBERG"
	
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
	path := fmt.Sprintf("%scatalogs/%s", models.ApiBase, catalogName)

	// Make the PUT request
	respBody, err := c.compaction.DoRequest(ctx, http.MethodPut, path, url.Values{}, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update catalog %s: %w", catalogName, err)
	}

	var resp models.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated catalog %s", catalogName),
	}, nil
}

// DeleteCatalog deletes a catalog from Amoro
func (c *Service) DeleteCatalog(ctx context.Context, catalogName string) (*models.CatalogResponse, error) {
	// Validate required fields
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}

	// Build the API path
	path := fmt.Sprintf("%scatalogs/%s", models.ApiBase, catalogName)

	// Make the DELETE request
	respBody, err := c.compaction.DoRequest(ctx, http.MethodDelete, path, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to delete catalog %s: %w", catalogName, err)
	}

	var resp models.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 && resp.Code != 0 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted catalog %s", catalogName),
	}, nil
}
