package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction"
	cmpClient "github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type AppService struct {
	db         *database.Database
	etl        *etl.ETLService
	compaction *compaction.CompactionService
}

func InitAppService(db *database.Database) (*AppService, error) {
	compactionClient := cmpClient.NewClient()

	// Initialize ETL service
	etlSvc, err := etl.InitAppService(db, compactionClient)
	if err != nil {
		return nil, err
	}

	// Initialize Compaction service
	compactionSvc, err := compaction.InitAppService(db, compactionClient)
	if err != nil {
		return nil, err
	}

	return &AppService{
		db:         db,
		etl:        etlSvc,
		compaction: compactionSvc,
	}, nil
}

func (s *AppService) ETL() *etl.ETLService {
	return s.etl
}

func (s *AppService) Compaction() *compaction.CompactionService {
	return s.compaction
}
