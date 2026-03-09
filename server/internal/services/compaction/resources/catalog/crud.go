package catalog

import (
	"context"
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

// returns the list of catalogs
func (s *Service) GetCatalogs(ctx context.Context) (interface{}, error) {
	path := models.ApiBase + "catalogs"
	result, err := s.compaction.Do(ctx, http.MethodGet, path, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all catalogs: %w", err)
	}

	return result, nil
}

// creates a new catalog
func (s *Service) CreateCatalog(ctx context.Context, req models.CatalogRequest) (*models.CatalogResponse, error) {
	if err := validateCatalog(&req); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%scatalogs", models.ApiBase)
	if err := s.compaction.DoAndValidate(ctx, http.MethodPost, path, url.Values{}, req); err != nil {
		return nil, fmt.Errorf("failed to create catalog %s: %w", req.Name, err)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully created catalog %s", req.Name),
	}, nil
}

// updates an existing catalog
func (s *Service) UpdateCatalog(ctx context.Context, catalogName string, req models.CatalogRequest) (*models.CatalogResponse, error) {
	if err := validateCatalog(&req); err != nil {
		return nil, err
	}

	req.Name = catalogName
	path := fmt.Sprintf("%scatalogs/%s", models.ApiBase, catalogName)
	if err := s.compaction.DoAndValidate(ctx, http.MethodPost, path, url.Values{}, req); err != nil {
		return nil, fmt.Errorf("failed to update catalog %s: %w", req.Name, err)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated catalog %s", catalogName),
	}, nil
}

// DeleteCatalog deletes a catalog from Amoro
func (s *Service) DeleteCatalog(ctx context.Context, catalogName string) (*models.CatalogResponse, error) {
	path := fmt.Sprintf("%scatalogs/%s", models.ApiBase, catalogName)
	if err := s.compaction.DoAndValidate(ctx, http.MethodDelete, path, url.Values{}, nil); err != nil {
		return nil, fmt.Errorf("failed to delete catalog %s: %w", catalogName, err)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("successfully deleted catalog %s", catalogName),
	}, nil
}

// validates the necessary requirements for creating or updating a catalog
func validateCatalog(req *models.CatalogRequest) error {
	if req.Name == "" {
		return fmt.Errorf("catalog name is required")
	}
	if req.Type == "" {
		return fmt.Errorf("catalog type is required")
	}

	if req.TableFormatList == nil || len(req.TableFormatList) == 0 {
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

	return nil
}
