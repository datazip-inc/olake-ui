package services

import (
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/temporal"
)

// AppService is a unified service exposing all domain operations backed by shared deps.
type AppService struct {
	// single ORM facade using one Ormer
	db         *database.Database
	tempClient *temporal.Client
}

// InitAppService constructs a unified AppService with singletons.
func InitAppService() (*AppService, error) {
	db, err := database.Init()
	if err != nil {
		return nil, err
	}
	return NewAppService(db)
}

// NewAppService wires a new AppService with singletons.
func NewAppService(db *database.Database) (*AppService, error) {
	client, err := temporal.NewClient()
	if err != nil {
		return nil, err
	}

	return &AppService{
		db:         db,
		tempClient: client,
	}, nil
}
