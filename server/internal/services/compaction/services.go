package compaction

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/catalog"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/table"
	aggregator "github.com/datazip-inc/olake-ui/server/internal/services/compaction/views"
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
