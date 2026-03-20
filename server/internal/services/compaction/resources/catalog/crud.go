package catalog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	olake "github.com/datazip-inc/olake/destination/iceberg"
)

type Service struct {
	compaction *client.Compaction
}

func NewService(c *client.Compaction) *Service {
	return &Service{
		compaction: c,
	}
}

func (s *Service) GetCatalogs(ctx context.Context) (interface{}, error) {
	path := models.APIBase + "catalogs"
	result, err := s.compaction.Do(ctx, http.MethodGet, path, url.Values{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all catalogs: %w", err)
	}

	return result, nil
}

func (s *Service) GetCatalog(ctx context.Context, catalogName string) (*models.CatalogRequest, error) {
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}

	path := fmt.Sprintf("%scatalogs/%s", models.APIBase, catalogName)
	var result models.CatalogRequest
	if err := s.compaction.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get catalog %s: %w", catalogName, err)
	}

	return &result, nil
}

func (s *Service) GetCatalogAsOLakeConfig(ctx context.Context, catalogName string) (*olake.Config, error) {
	catalog, err := s.GetCatalog(ctx, catalogName)
	if err != nil {
		return nil, err
	}

	return MapCompactionCatalogToOLakeConfig(catalog)
}

// creates a new catalog
func (s *Service) CreateCatalog(ctx context.Context, req *models.CatalogRequest) (*models.CatalogResponse, error) {
	if err := validateCatalog(req); err != nil {
		return nil, err
	}

	// Set default table properties for Iceberg tables
	setDefaultTableProperties(req)

	path := fmt.Sprintf("%scatalogs", models.APIBase)
	if err := s.compaction.DoAndValidate(ctx, http.MethodPost, path, url.Values{}, req); err != nil {
		return nil, fmt.Errorf("failed to create catalog %s: %w", req.Name, err)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully created catalog %s", req.Name),
	}, nil
}

func (s *Service) CreateCatalogFromOLakeConfig(ctx context.Context, configJSON string) (*models.CatalogResponse, error) {
	catalogReq, err := MapOLakeConfigToCompactionCatalog(configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to map OLake config to catalog: %w", err)
	}

	return s.CreateCatalog(ctx, catalogReq)
}

// updates an existing catalog
func (s *Service) UpdateCatalog(ctx context.Context, catalogName string, req *models.CatalogRequest) (*models.CatalogResponse, error) {
	if err := validateCatalog(req); err != nil {
		return nil, err
	}

	req.Name = catalogName
	path := fmt.Sprintf("%scatalogs/%s", models.APIBase, catalogName)
	if err := s.compaction.DoAndValidate(ctx, http.MethodPut, path, url.Values{}, req); err != nil {
		return nil, fmt.Errorf("failed to update catalog %s: %w", req.Name, err)
	}

	return &models.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated catalog %s", catalogName),
	}, nil
}

func (s *Service) UpdateCatalogFromOLakeConfig(ctx context.Context, configJSON string) (*models.CatalogResponse, error) {
	catalogReq, err := MapOLakeConfigToCompactionCatalog(configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to map OLake config to catalog: %w", err)
	}

	return s.UpdateCatalog(ctx, catalogReq.Name, catalogReq)
}

// DeleteCatalog deletes a catalog from Amoro
func (s *Service) DeleteCatalog(ctx context.Context, catalogName string) (*models.CatalogResponse, error) {
	path := fmt.Sprintf("%scatalogs/%s", models.APIBase, catalogName)
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

	if len(req.TableFormatList) == 0 {
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

func setDefaultTableProperties(req *models.CatalogRequest) {
	if req.Properties == nil {
		req.Properties = make(map[string]string)
	}

	req.Properties["table.self-optimizing.enabled"] = "false"
	req.Properties["table.self-optimizing.quota"] = "0.1"
}

// SetCatalogTableProperty sets a table property in the catalog's tableProperties map
// The key format is: <database>:<table>
// The value format is: <enabled>,<minor_interval>,<major_interval>,<full_interval>
func (s *Service) SetCatalogTableProperty(ctx context.Context, catalogName, database, table, _, value string) (*models.CatalogResponse, error) {
	catalogReq, err := s.GetCatalog(ctx, catalogName)
	if err != nil {
		return nil, err
	}

	if catalogReq.TableProperties == nil {
		catalogReq.TableProperties = make(map[string]string)
	}

	// Use <database>:<table> as the key
	tableKey := fmt.Sprintf("%s:%s", database, table)
	catalogReq.TableProperties[tableKey] = value

	return s.UpdateCatalog(ctx, catalogName, catalogReq)
}
