package services

import (
	"os"

	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction"
	cmpClient "github.com/datazip-inc/olake-ui/server/internal/services/compaction/client"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type AppService struct {
	db         *database.Database
	etl        *etl.ETLService
	compaction *compaction.CompactionService
	amoroClient *cmpClient.Compaction
}

func InitAppService(db *database.Database) (*AppService, error) {
	// Create Amoro client once
	baseURL := os.Getenv("AMORO_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:1630"
	}

	apiKey := os.Getenv("AMORO_API_KEY")
	apiSecret := os.Getenv("AMORO_API_SECRET")
	amoroClient := cmpClient.NewClient(baseURL, apiKey, apiSecret)

	// Initialize ETL service with shared Amoro client
	etlSvc, err := etl.InitAppService(db, amoroClient)
	if err != nil {
		return nil, err
	}

	// Initialize Compaction service with shared Amoro client
	compactionSvc, err := compaction.InitCompactionService(db, amoroClient)
	if err != nil {
		return nil, err
	}

	return &AppService{
		db:          db,
		etl:         etlSvc,
		compaction:  compactionSvc,
		amoroClient: amoroClient,
	}, nil
}

func (s *AppService) ETL() *etl.ETLService {
	return s.etl
}

func (s *AppService) Compaction() *compaction.CompactionService {
	return s.compaction
}
