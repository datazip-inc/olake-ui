package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction"
	cmpClient "github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type AppService struct {
	db         *database.Database
	etl        *etl.Service
	compaction *compaction.Service
}

func InitAppService(db *database.Database) (*AppService, error) {
	compactionClient := cmpClient.NewClient()

	// Initialize ETL service
	etlSvc, err := etl.InitService(db, compactionClient)
	if err != nil {
		return nil, err
	}

	// Initialize Compaction service
	compactionSvc, err := compaction.InitService(db, compactionClient)
	if err != nil {
		return nil, err
	}

	return &AppService{
		db:         db,
		etl:        etlSvc,
		compaction: compactionSvc,
	}, nil
}

func (s *AppService) ETL() *etl.Service {
	return s.etl
}

func (s *AppService) Compaction() *compaction.Service {
	return s.compaction
}
