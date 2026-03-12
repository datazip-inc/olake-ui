package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type AppService struct {
	db  *database.Database
	etl *etl.Service
}

func InitAppService(db *database.Database) (*AppService, error) {
	// Initialize ETL service
	etlSvc, err := etl.InitService(db)
	if err != nil {
		return nil, err
	}

	return &AppService{
		db:  db,
		etl: etlSvc,
	}, nil
}

func (s *AppService) ETL() *etl.Service {
	return s.etl
}
