package services

import (
	"context"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/services/optimisation"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

type AppService struct {
	db  *database.Database
	etl *etl.Service
	opt *optimisation.Service
}

func InitAppService(db *database.Database) (*AppService, error) {
	// Initialize ETL service
	etlSvc, err := etl.InitService(db)
	if err != nil {
		return nil, err
	}

	var optSvc *optimisation.Service
	enableOptimisation := web.AppConfig.DefaultBool(constants.ConfEnableOptimisation, false)
	if enableOptimisation {
		optSvc, err = optimisation.InitService()
		if err != nil {
			return nil, err
		}
	}

	appSvc := &AppService{
		db:  db,
		etl: etlSvc,
		opt: optSvc,
	}

	if enableOptimisation {
		appSvc.CreateAllDestAsCatalogs(context.Background())
	}

	return appSvc, nil
}

func (s *AppService) ETL() *etl.Service {
	return s.etl
}

func (s *AppService) Optimisation() *optimisation.Service {
	return s.opt
}

func (s *AppService) CreateAllDestAsCatalogs(ctx context.Context) {
	destinations, err := s.db.ListDestinations()
	if err != nil {
		logger.Errorf("Failed to list destinations for optimisation sync: %v", err)
		return
	}

	for _, dest := range destinations {
		if !strings.EqualFold(dest.DestType, "iceberg") {
			continue
		}

		if _, err := s.opt.CreateCatalogFromOLakeConfig(ctx, dest.Config); err != nil {
			logger.Errorf("Failed to create catalog: %s", err)
		} else {
			logger.Infof("Catalog created successfully")
		}
	}
}
