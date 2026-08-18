package etl

import (
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

// AppService is a unified service exposing all domain operations backed by shared deps.
type Service struct {
	// single ORM facade using one Ormer
	db       *database.Database
	temporal *temporal.Temporal
	metrics  *metricsCollector
}

// InitAppService constructs a unified AppService with singletons.
func InitService(db *database.Database) (*Service, error) {
	client, err := temporal.NewClient()
	if err != nil {
		return nil, err
	}

	return &Service{
		db:       db,
		temporal: client,
		metrics:  newMetricsCollector(db, client, constants.DefaultConfigDir),
	}, nil
}
