package compaction

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/views"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/catalogs"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/resources/tables"
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

func InitCompactionService(db *database.Database, amoroClient *client.Compaction) (*CompactionService, error) {
	return &CompactionService{
		db:     db,
		client: amoroClient,
	}, nil
}

func (s *CompactionService) GetClient() *client.Compaction {
	return s.client
}

func (s *CompactionService) GetDB() *database.Database {
	return s.db
}
