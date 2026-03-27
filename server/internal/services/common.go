package services

import (
	"context"
	"sort"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/optimization"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// DestinationWithCatalogResult carries the outcome of a destination operation that includes
// an optional catalog side-effect. CatalogErr is non-nil when the ETL operation succeeded
// but the catalog operation failed.
type DestinationWithCatalogResult struct {
	CatalogErr error
}

func (s *AppService) CreateDestinationWithCatalog(ctx context.Context, projectID string, req *dto.CreateDestinationRequest, userID *int) (*DestinationWithCatalogResult, error) {
	if err := s.etl.CreateDestination(ctx, req, projectID, userID); err != nil {
		return nil, err
	}

	result := &DestinationWithCatalogResult{}
	if s.opt != nil {
		catalogName, _ := optimization.ExtractCatalogNameFromConfig(req.Config)
		logger.Debugf("Creating catalog[%s] from destination[%s]", catalogName, req.Name)
		if _, err := s.opt.CreateCatalogFromOLakeConfig(ctx, req.Config, true); err != nil {
			logger.Errorf("Failed to create catalog[%s]: %s", catalogName, err)
			result.CatalogErr = err
		} else {
			logger.Infof("Catalog[%s] created successfully", catalogName)
		}
	}
	return result, nil
}

func (s *AppService) UpdateDestinationWithCatalog(ctx context.Context, id int, projectID string, req *dto.UpdateDestinationRequest, userID *int) (*DestinationWithCatalogResult, error) {
	if err := s.etl.UpdateDestination(ctx, id, projectID, req, userID); err != nil {
		return nil, err
	}

	result := &DestinationWithCatalogResult{}
	if s.opt != nil && req.Config != "" {
		catalogName, _ := optimization.ExtractCatalogNameFromConfig(req.Config)
		logger.Debugf("Updating catalog[%s] from destination[%s]", catalogName, req.Name)
		if _, err := s.opt.UpdateCatalogFromOLakeConfig(ctx, req.Config); err != nil {
			logger.Errorf("Failed to update catalog[%s]: %s", catalogName, err)
			result.CatalogErr = err
		} else {
			logger.Infof("Catalog[%s] updated successfully", catalogName)
		}
	}
	return result, nil
}

func (s *AppService) DeleteDestinationWithCatalog(ctx context.Context, id int) (*dto.DeleteDestinationResponse, *DestinationWithCatalogResult, error) {
	var catalogName string
	if s.opt != nil {
		if destination, err := s.etl.GetDestinationByID(id); err == nil {
			catalogName, _ = optimization.ExtractCatalogNameFromConfig(destination.Config)
		}
	}

	resp, err := s.etl.DeleteDestination(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	result := &DestinationWithCatalogResult{}
	if s.opt != nil && catalogName != "" {
		logger.Debugf("Deleting catalog[%s]", catalogName)
		if _, err := s.opt.DeleteCatalog(ctx, catalogName); err != nil {
			logger.Errorf("Failed to delete catalog[%s]: %s", catalogName, err)
			result.CatalogErr = err
		} else {
			logger.Infof("Catalog[%s] deleted successfully", catalogName)
		}
	}
	return resp, result, nil
}

func (s *AppService) SyncCatalogs(ctx context.Context) {
	destinations, err := s.db.ListDestinations()
	if err != nil {
		logger.Errorf("Failed to list destinations for optimization sync: %s", err)
		return
	}

	// Sort destinations by "CreatedAt" descending (newest first)
	sort.Slice(destinations, func(i, j int) bool {
		return destinations[i].CreatedAt.After(destinations[j].CreatedAt)
	})

	logger.Debugf("Syncing catalogs for %d destinations in descending order of creation", len(destinations))

	for _, dest := range destinations {
		if !strings.EqualFold(dest.DestType, "iceberg") {
			logger.Debugf("Skipping catalog creation for destination [%s] as the type is not iceberg", dest.Name)
			continue
		}

		catalogName, _ := optimization.ExtractCatalogNameFromConfig(dest.Config)
		logger.Debugf("Creating catalog for destination[%s] catalog[%s]", dest.Name, catalogName)

		if _, err := s.opt.CreateCatalogFromOLakeConfig(ctx, dest.Config, true); err != nil {
			logger.Errorf("Failed to create catalog[%s] from destination[%s]: %s", catalogName, dest.Name, err)
		} else {
			logger.Infof("Catalog[%s] created successfully from destination[%s]", catalogName, dest.Name)
		}
	}
}
