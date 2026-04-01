package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
)

type ETLService struct {
	db       *database.Database
	temporal *temporal.Temporal
	RBAC     *RBACService // ← add
}

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
