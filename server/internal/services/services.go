package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type AppService struct {
	db         *database.Database
	etl        *etl.ETLService
	compaction *compaction.CompactionService
}

func InitAppService(db *database.Database) (*AppService, error) {
	etlSvc, err := etl.InitAppService(db)
	if err != nil {
		return nil, err
	}

	compactionSvc, err := compaction.InitCompactionService(db)
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

func (s *AppService) DB() *database.Database {
	return s.db
}
