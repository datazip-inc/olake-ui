package services

import (
	"context"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/services/optimization"
)

type AppService struct {
	db  *database.Database
	etl *etl.Service
	opt *optimization.Service
}

func InitAppService(db *database.Database) (*AppService, error) {
	// Initialize ETL service
	etlSvc, err := etl.InitService(db)
	if err != nil {
		return nil, err
	}

	appSvc := &AppService{
		db:  db,
		etl: etlSvc,
		opt: nil,
	}
	// TODO BEFORE MERGE
	// enableOptimization := web.AppConfig.DefaultBool(constants.ConfEnableOptimization, false)
	enableOptimization := true
	if enableOptimization {
		optSvc, err := optimization.InitService()
		if err != nil {
			return nil, err
		}

		appSvc.opt = optSvc

		// TODO: define context in main and pass
		appSvc.SyncCatalogs(context.Background())
	}

	return appSvc, nil
}

func (s *AppService) ETL() *etl.Service {
	return s.etl
}

func (s *AppService) Optimization() *optimization.Service {
	return s.opt
}
