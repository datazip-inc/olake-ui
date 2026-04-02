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

func (s *Service) GetCatalog(ctx context.Context, catalogName string) (*models.Config, error) {
	catalog, err := s.getCatalogInOpt(ctx, catalogName)
	if err != nil {
		return nil, err
	}

	// map the catalog details received from opt
	// to destination config
	return mapCatalogToDest(catalog)
}

func (s *Service) getCatalogInOpt(ctx context.Context, catalogName string) (*dto.CatalogRequest, error) {
	path := fmt.Sprintf(constants.OptPathCatalogDetail, catalogName)

	var result dto.CatalogRequest
	if err := s.DoInto(ctx, http.MethodGet, path, url.Values{}, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get catalog %s: %s", catalogName, err)
	}

	return &result, nil
}

func (s *Service) CreateCatalog(ctx context.Context, configJSON string) (string, error) {
	req, err := s.createOptConfig(configJSON, false)
	if err != nil {
		return "", fmt.Errorf("failed to create optimization config: %s", err)
	}

	if err := validateCatalog(req); err != nil {
		return "", fmt.Errorf("failed to validate catalog config in optimization: %s", err)
	}

	// set default catalog properties
	setDefaultCatalogProperties(req)

	path := constants.OptPathCatalogs
	if err := s.DoExec(ctx, http.MethodPost, path, url.Values{}, req); err != nil {
		return "", fmt.Errorf("failed to create catalog %s: %s", req.Name, err)
	}

	return req.Name, nil
}

func (s *Service) UpdateCatalog(ctx context.Context, configJSON string) (string, error) {
	req, err := s.createOptConfig(configJSON, true)
	if err != nil {
		return "", fmt.Errorf("failed to create optimization config: %s", err)
	}

	if err := validateCatalog(req); err != nil {
		return "", fmt.Errorf("failed to validate catalog config in optimization: %s", err)
	}

	existing, err := s.getCatalogInOpt(ctx, req.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get existing catalog %s for update: %s", req.Name, err)
	}

	req.Properties = mergeMaps(existing.Properties, req.Properties)
	req.StorageConfig = mergeMaps(existing.StorageConfig, req.StorageConfig)
	req.AuthConfig = mergeMaps(existing.AuthConfig, req.AuthConfig)
	req.TableProperties = mergeMaps(existing.TableProperties, req.TableProperties)

	path := fmt.Sprintf(constants.OptPathCatalogDetail, req.Name)
	if err := s.DoExec(ctx, http.MethodPut, path, url.Values{}, req); err != nil {
		return "", fmt.Errorf("failed to update catalog %s in optimization: %s", req.Name, err)
	}

	return req.Name, nil
}

func (s *Service) DeleteCatalog(ctx context.Context, catalogName string) (string, error) {
	path := fmt.Sprintf(constants.OptPathCatalogDetail, catalogName)
	if err := s.DoExec(ctx, http.MethodDelete, path, url.Values{}, nil); err != nil {
		return "", fmt.Errorf("failed to delete catalog %s: %s", catalogName, err)
	}

	return catalogName, nil
}

func (s *Service) createOptConfig(configJSON string, update bool) (*dto.CatalogRequest, error) {
	var config models.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse ETL config: %s", err)
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
		Name:                    config.CatalogName,
		Type:                    catalogType,
		OptimizerGroup:          og,
		OptimizeTableFormatList: constants.OptimizeTableFormatList,
		StorageConfig:           make(map[string]string),
		AuthConfig:              make(map[string]string),
		Properties:              make(map[string]string),
		TableProperties:         make(map[string]string),
	}

	optimizationReq.StorageConfig["storage.type"] = constants.DefaultOptimizationStorageType

	mapAuthConfig(&config, optimizationReq.AuthConfig, optimizationReq.StorageConfig)
	mapCatalogProperties(&config, optimizationReq.Properties, string(config.CatalogType))
	if !update {
		setDefaultCatalogProperties(optimizationReq)
	}

	return optimizationReq, nil
}

// validates the necessary requirements for creating or updating a catalog
func validateCatalog(req *dto.CatalogRequest) error {
	if req.Name == "" {
		return fmt.Errorf("catalog name is required")
	}
	if req.Type == "" {
		return fmt.Errorf("catalog type is required")
	}

	if len(req.OptimizeTableFormatList) == 0 {
		req.OptimizeTableFormatList = []string{"ICEBERG"}
	}

	return nil
}
