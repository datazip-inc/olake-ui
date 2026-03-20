package services

import (
	"fmt"

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

func InitAppService(db *database.Database, enableCompaction bool) (*AppService, error) {
	var compactionClient *cmpClient.Compaction
	var compactionSvc *compaction.Service
	var err error

	if enableCompaction {
		compactionClient, err = cmpClient.NewClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create compaction client: %s", err)
		}

		// Initialize Compaction service
		compactionSvc, err = compaction.InitService(db, compactionClient)
		if err != nil {
			return nil, err
		}
	}

	// Initialize ETL service
	etlSvc, err := etl.InitService(db, compactionClient)
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

func (s *AppService) IsCompactionEnabled() bool {
	return s.compaction != nil
}
