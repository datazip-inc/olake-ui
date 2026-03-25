package services

import (
	"context"
	"sort"
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


	appSvc := &AppService{
		db:  db,
		etl: etlSvc,
		opt: nil,
	}
	enableOptimisation := web.AppConfig.DefaultBool(constants.ConfEnableOptimisation, false)
	if enableOptimisation {
		optSvc, err := optimisation.InitService()
		if err != nil {
			return nil, err
		}

		appSvc.opt = optSvc
		appSvc.SyncCatalog(context.Background())
	}

	return appSvc, nil
}

func (s *AppService) ETL() *etl.Service {
	return s.etl
}

func (s *AppService) Optimisation() *optimisation.Service {
	return s.opt
}

func (s *AppService) SyncCatalog(ctx context.Context) {
	destinations, err := s.db.ListDestinations()
	if err != nil {
		logger.Errorf("Failed to list destinations for optimisation sync: %s", err)
		return
	}

	// Sort destinations by "CreatedAt" descending (newest first)
	sort.Slice(destinations, func(i, j int) bool {
		return destinations[i].CreatedAt.After(destinations[j].CreatedAt)
	})

	logger.Debugf("Syncing catalogs for %d destinations in descending order of creation", len(destinations))

	for _, dest := range destinations {
		if !strings.EqualFold(dest.DestType, "iceberg") {
			logger.Debugf("Skipping catalog creation for destination [%s] as the type is not iceberg", dest.Name)
			continue
		}

		catalogName, _ := optimisation.ExtractCatalogNameFromConfig(dest.Config)
		logger.Debugf("Creating catalog for destination[%s] catalog[%s]", dest.Name, catalogName)

		if _, err := s.opt.CreateCatalogFromOLakeConfig(ctx, dest.Config, true); err != nil {
			logger.Errorf("Failed to create catalog[%s] from destination[%s]: %s", catalogName, dest.Name, err)
		} else {
			logger.Infof("Catalog[%s] created successfully from destination[%s]", catalogName, dest.Name)
		}
	}
}
