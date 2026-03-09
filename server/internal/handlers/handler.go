package handlers

import (
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type Handler struct {
	etl      *services.ETLService
	sessions *sessionStore
}

func NewGinHandler(s *services.ETLService, cfg appconfig.Config) (*Handler, error) {
	sessionStore, err := newSessionStore(cfg)
	if err != nil {
		return nil, err
	}
	return &Handler{
		etl:      s,
		sessions: sessionStore,
	}, nil
}
