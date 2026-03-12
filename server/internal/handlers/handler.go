package handlers

import (
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type Handler struct {
	etl      *services.ETLService
	sessions *sessionStore
}

func NewHandler(s *services.ETLService, cfg *appconfig.Config, db *database.Database) (*Handler, error) {
	sessionStore, err := newSessionStore(cfg, db)
	if err != nil {
		return nil, err
	}
	return &Handler{
		etl:      s,
		sessions: sessionStore,
	}, nil
}
