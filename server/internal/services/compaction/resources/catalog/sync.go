package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

func (s *Service) SyncCatalogToFusion(ctx context.Context, configJSON string, isUpdate bool) error {
	CompactionCatalogReq, err := MapETLConfigToCompactionCatalog(configJSON)
	if err != nil {
		return fmt.Errorf("failed to map OLake config to Compaction catalog: %w", err)
	}

	// Mark this catalog as created from ETL
	if CompactionCatalogReq.Properties == nil {
		CompactionCatalogReq.Properties = make(map[string]string)
	}
	CompactionCatalogReq.Properties["olake_created"] = "true"

	logger.Infof("Syncing catalog to Compaction: name=%s, type=%s, isUpdate=%v", CompactionCatalogReq.Name, CompactionCatalogReq.Type, isUpdate)

	if isUpdate {
		if err := s.updateCatalogInCompaction(ctx, CompactionCatalogReq); err != nil {
			logger.Errorf("Failed to update catalog in Compaction: %v", err)
			return fmt.Errorf("failed to update catalog in Compaction: %w", err)
		}
		logger.Infof("Successfully updated catalog %s in Compaction", CompactionCatalogReq.Name)
	} else {
		if err := s.createCatalogInCompaction(ctx, CompactionCatalogReq); err != nil {
			logger.Errorf("Failed to create catalog in Compaction: %v", err)
			return fmt.Errorf("failed to create catalog in Compaction: %w", err)
		}
		logger.Infof("Successfully created catalog %s in Compaction", CompactionCatalogReq.Name)
	}

	return nil
}

func (s *Service) createCatalogInCompaction(ctx context.Context, req *models.CatalogRequest) error {
	resp, err := s.CreateCatalog(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("catalog creation failed: %s", resp.Message)
	}

	return nil
}

func (s *Service) updateCatalogInCompaction(ctx context.Context, req *models.CatalogRequest) error {
	catalogExists, err := s.CheckCatalogExists(ctx, req.Name)
	if err != nil {
		logger.Warnf("Failed to check if catalog exists in Compaction: %v", err)
	}

	if !catalogExists {
		return nil
	}

	resp, err := s.UpdateCatalog(ctx, req.Name, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("catalog update failed: %s", resp.Message)
	}

	return nil
}

func (s *Service) CheckCatalogExists(ctx context.Context, catalogName string) (bool, error) {
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
