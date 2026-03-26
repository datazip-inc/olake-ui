package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

func (s *Service) GetCatalog(ctx context.Context, catalogName string) (*dto.CatalogRequest, error) {
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}

	path := fmt.Sprintf("%scatalogs/%s", constants.OptimizationAPIBase, catalogName)
	var result dto.CatalogRequest
	if err := s.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get catalog %s: %s", catalogName, err)
	}

	return &result, nil
}

// creates a new catalog
func (s *Service) CreateCatalog(ctx context.Context, req *dto.CatalogRequest) (*dto.CatalogResponse, error) {
	if err := validateCatalog(req); err != nil {
		return nil, err
	}

	// Set default table properties for Iceberg tables
	setDefaultCatalogProperties(req)

	path := fmt.Sprintf("%scatalogs", constants.OptimizationAPIBase)
	if err := s.DoAndValidate(ctx, http.MethodPost, path, url.Values{}, req); err != nil {
		return nil, fmt.Errorf("failed to create catalog %s: %s", req.Name, err)
	}

	return &dto.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully created catalog %s", req.Name),
	}, nil
}

// updates an existing catalog
func (s *Service) UpdateCatalog(ctx context.Context, catalogName string, req *dto.CatalogRequest) (*dto.CatalogResponse, error) {
	if err := validateCatalog(req); err != nil {
		return nil, err
	}

	req.Name = catalogName
	path := fmt.Sprintf("%scatalogs/%s", constants.OptimizationAPIBase, catalogName)
	if err := s.DoAndValidate(ctx, http.MethodPut, path, url.Values{}, req); err != nil {
		return nil, fmt.Errorf("failed to update catalog %s: %s", req.Name, err)
	}

	return &dto.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated catalog %s", catalogName),
	}, nil
}

func (s *Service) UpdateCatalogFromOLakeConfig(ctx context.Context, configJSON string) (*dto.CatalogResponse, error) {
	catalogReq, err := CreateOptimizationConfig(configJSON, true)
	if err != nil {
		return nil, fmt.Errorf("failed to map OLake config to catalog: %s", err)
	}

	return s.UpdateCatalog(ctx, catalogReq.Name, catalogReq)
}

func (s *Service) GetCatalogAsOLakeConfig(ctx context.Context, catalogName string) (*models.Config, error) {
	catalog, err := s.GetCatalog(ctx, catalogName)
	if err != nil {
		return nil, err
	}

	return MapoptimizationCatalogToOLakeConfig(catalog)
}

func (s *Service) CreateCatalogFromOLakeConfig(ctx context.Context, configJSON string, byETL bool) (*dto.CatalogResponse, error) {
	catalogReq, err := CreateOptimizationConfig(configJSON, false)
	if err != nil {
		return nil, fmt.Errorf("failed to map OLake config to catalog: %s", err)
	}

	if byETL {
		catalogReq.Properties["byOLake"] = "true"
	} else {
		catalogReq.Properties["byOLake"] = "false"
	}

	return s.CreateCatalog(ctx, catalogReq)
}

// ExtractCatalogNameFromConfig extracts catalog name from destination config JSON
func ExtractCatalogNameFromConfig(configJSON string) (string, error) {
	var config models.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return "", fmt.Errorf("failed to parse config: %s", err)
	}

	if config.CatalogName == "" {
		return "", fmt.Errorf("catalog_name not found in config")
	}

	return config.CatalogName, nil
}

func CreateOptimizationConfig(configJSON string, update bool) (*dto.CatalogRequest, error) {
	var config models.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse OLake config: %s", err)
	}

	if config.CatalogName == "" {
		return nil, fmt.Errorf("catalog_name is required in config")
	}

	og, err := web.AppConfig.String(constants.ConfOptimizationGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization group")
	}

	catalogType := normalizeCatalogType(string(config.CatalogType))
	optimizationReq := &dto.CatalogRequest{
		Name:            config.CatalogName,
		Type:            catalogType,
		OptimizerGroup:  og,
		TableFormatList: constants.TableFormatList,
		StorageConfig:   make(map[string]string),
		AuthConfig:      make(map[string]string),
		Properties:      make(map[string]string),
		TableProperties: make(map[string]string),
	}

	optimizationReq.StorageConfig["storage.type"] = constants.DefaultStroageType

	mapAuthConfig(&config, optimizationReq.AuthConfig, optimizationReq.StorageConfig)
	mapCatalogProperties(&config, optimizationReq.Properties, string(config.CatalogType))
	if update == false {
		setDefaultCatalogProperties(optimizationReq)
	}

	return optimizationReq, nil
}

// deletes an existing catalog
func (s *Service) DeleteCatalog(ctx context.Context, catalogName string) (*dto.CatalogResponse, error) {
	if catalogName == "" {
		return nil, fmt.Errorf("catalog name is required")
	}

	path := fmt.Sprintf("%scatalogs/%s", constants.OptimizationAPIBase, catalogName)
	if err := s.DoAndValidate(ctx, http.MethodDelete, path, url.Values{}, nil); err != nil {
		return nil, fmt.Errorf("failed to delete catalog %s: %s", catalogName, err)
	}

	return &dto.CatalogResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted catalog %s", catalogName),
	}, nil
}

// validates the necessary requirements for creating or updating a catalog
func validateCatalog(req *dto.CatalogRequest) error {
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
