package compaction

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/catalog"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/table"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/views"
)

// CompactionService is a unified service for compaction operations
type CompactionService struct {
	db           *database.Database
	client       *client.Compaction
	Catalog      *catalog.Service
	Table        *table.Service
	Optimization *optimization.Service
	Aggregator   *aggregator.Service
}

func InitAppService(db *database.Database, client *client.Compaction) (*CompactionService, error) {
	catalogSvc := catalog.NewService(client)
	tableSvc := table.NewService(client)

	return &CompactionService{
		db:           db,
		client:       client,
		Catalog:      catalogSvc,
		Table:        tableSvc,
		Optimization: optimization.NewService(client, tableSvc),
		Aggregator:   aggregator.NewService(client, catalogSvc, tableSvc),
	}, nil
}

func (s *CompactionService) GetClient() *client.Compaction {
	return s.client
}

func (s *CompactionService) GetDB() *database.Database {
	return s.db
}
