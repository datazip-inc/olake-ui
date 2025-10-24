package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/temporal"
)

// AppService is a unified service exposing all domain operations backed by shared deps.
type ETLService struct {
	// single ORM facade using one Ormer
	db       *database.Database
	temporal *temporal.Temporal
}

// InitAppService constructs a unified AppService with singletons.
func InitAppService(db *database.Database) (*ETLService, error) {
	client, err := temporal.NewClient()
	if err != nil {
		return nil, err
	}

	return &ETLService{
		db:       db,
		temporal: client,
	}, nil
}
