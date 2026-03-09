package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

func (s *Service) SyncCatalogToFusion(ctx context.Context, destinationName, destinationType, configJSON string, isUpdate bool) error {
	amoroCatalogReq, err := MapOLakeConfigToCompactionCatalog(destinationName, configJSON)
	if err != nil {
		return fmt.Errorf("failed to map OLake config to Amoro catalog: %w", err)
	}

	logger.Infof("Syncing catalog to Amoro: name=%s, type=%s, isUpdate=%v", amoroCatalogReq.Name, amoroCatalogReq.Type, isUpdate)

	if isUpdate {
		if err := s.updateCatalogInAmoro(ctx, amoroCatalogReq); err != nil {
			logger.Errorf("Failed to update catalog in Amoro: %v", err)
			return fmt.Errorf("failed to update catalog in Amoro: %w", err)
		}
		logger.Infof("Successfully updated catalog %s in Amoro", amoroCatalogReq.Name)
	} else {
		if err := s.createCatalogInAmoro(ctx, amoroCatalogReq); err != nil {
			logger.Errorf("Failed to create catalog in Amoro: %v", err)
			return fmt.Errorf("failed to create catalog in Amoro: %w", err)
		}
		logger.Infof("Successfully created catalog %s in Amoro", amoroCatalogReq.Name)
	}

	return nil
}

func (s *Service) createCatalogInAmoro(ctx context.Context, req *models.CatalogRequest) error {
	catalogExists, err := s.checkCatalogExists(ctx, req.Name)
	if err != nil {
		logger.Warnf("Failed to check if catalog exists in Amoro: %v", err)
	}

	// in case the user had earlier created a catalog of the same name in fusion, and tries to import
	// a destination config from OLake of the same name to fusion
	if catalogExists {
		return fmt.Errorf("catalog with the same name already exists")
	}

	resp, err := s.CreateCatalog(ctx, *req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("catalog creation failed: %s", resp.Message)
	}

	return nil
}

func (s *Service) updateCatalogInAmoro(ctx context.Context, req *models.CatalogRequest) error {
	catalogExists, err := s.checkCatalogExists(ctx, req.Name)
	if err != nil {
		logger.Warnf("Failed to check if catalog exists in Amoro: %v", err)
	}

	if !catalogExists {
		return nil
	}

	resp, err := s.UpdateCatalog(ctx, req.Name, *req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("catalog update failed: %s", resp.Message)
	}

	return nil
}

func (s *Service) checkCatalogExists(ctx context.Context, catalogName string) (bool, error) {
	catalogs, err := s.GetCatalogs(ctx)
	if err != nil {
		return false, err
	}

	catalogsJSON, err := json.Marshal(catalogs)
	if err != nil {
		return false, err
	}

	var catalogList []map[string]interface{}
	if err := json.Unmarshal(catalogsJSON, &catalogList); err != nil {
		return false, err
	}

	for _, catalog := range catalogList {
		if name, ok := catalog["catalogName"].(string); ok && name == catalogName {
			return true, nil
		}
	}

	return false, nil
}
