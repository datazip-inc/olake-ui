package compaction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/catalog"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/table"
	aggregator "github.com/datazip-inc/olake-ui/server/internal/services/compaction/views"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// Service is a unified service for compaction operations
type Service struct {
	db           *database.Database
	client       *client.Compaction
	Catalog      *catalog.Service
	Table        *table.Service
	Optimization *optimization.Service
	Aggregator   *aggregator.Service
}

func InitService(db *database.Database, client *client.Compaction) (*Service, error) {
	catalogSvc := catalog.NewService(client)
	tableSvc := table.NewService(client)

	svc := &Service{
		db:           db,
		client:       client,
		Catalog:      catalogSvc,
		Table:        tableSvc,
		Optimization: optimization.NewService(client, tableSvc),
		Aggregator:   aggregator.NewService(client, catalogSvc, tableSvc),
	}

	// Sync all existing destinations at startup
	logger.Info("Syncing all destinations with compaction service...")
	svc.SyncAllDestinations(context.Background())
	logger.Info("Destination sync completed")

	return svc, nil
}

func (s *Service) GetClient() *client.Compaction {
	return s.client
}

func (s *Service) GetDB() *database.Database {
	return s.db
}

// SyncAllDestinations is called at startup to upsert all existing iceberg destinations as catalogs in compaction.
func (s *Service) SyncAllDestinations(ctx context.Context) {
	destinations, err := s.db.ListDestinations()
	if err != nil {
		logger.Errorf("Failed to list destinations for compaction sync: %v", err)
		return
	}

	catalogSvc := catalog.NewService(s.client)
	for _, dest := range destinations {
		if !strings.EqualFold(dest.DestType, "iceberg") {
			continue
		}

		// Check if catalog already exists using catalog_name from config
		catalogName, err := extractCatalogNameFromConfig(dest.Config)
		if err != nil {
			logger.Errorf("Failed to extract catalog_name from destination %s config: %v", dest.Name, err)
			continue
		}

		exists, err := catalogSvc.CheckCatalogExists(ctx, catalogName)
		if err != nil {
			logger.Warnf("Failed to check catalog existence for %s: %v", catalogName, err)
		}

		if err := catalogSvc.SyncCatalogToFusion(ctx, dest.Config, exists); err != nil {
			logger.Errorf("Failed to sync catalog %s for destination %s: %v", catalogName, dest.Name, err)
		} else {
			logger.Infof("Synced catalog %s for destination %s", catalogName, dest.Name)
		}
	}
}

// extractCatalogNameFromConfig extracts the catalog_name field from the config JSON
func extractCatalogNameFromConfig(configJSON string) (string, error) {
	var config models.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return "", fmt.Errorf("failed to parse config: %w", err)
	}

	if config.CatalogName == "" {
		return "", fmt.Errorf("catalog_name is required in config")
	}

	return config.CatalogName, nil
}
