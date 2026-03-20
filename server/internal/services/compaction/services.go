package compaction

import (
	"context"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
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

	return &Service{
		db:           db,
		client:       client,
		Catalog:      catalogSvc,
		Table:        tableSvc,
		Optimization: optimization.NewService(client, tableSvc),
		Aggregator:   aggregator.NewService(client, catalogSvc, tableSvc),
	}, nil
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
		if err := catalogSvc.UpsertCatalogInFusion(ctx, dest.Name, dest.Config); err != nil {
			logger.Errorf("Failed to upsert catalog for destination %s: %v", dest.Name, err)
		} else {
			logger.Infof("Synced catalog for destination %s", dest.Name)
		}
	}
}
