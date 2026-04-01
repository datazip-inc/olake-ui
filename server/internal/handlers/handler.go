package handlers

import (
	"github.com/casbin/casbin/v3"
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type Handler struct {
	etl      *services.ETLService
	sessions *sessionStore
	enforcer *casbin.Enforcer
}

func NewHandler(s *services.ETLService, cfg *appconfig.Config, enforcer *casbin.Enforcer) (*Handler, error) {
	sessionStore, err := newSessionStore(cfg)
	if err != nil {
		return nil, err
	}
	return &Handler{etl: s, sessions: sessionStore, enforcer: enforcer}, nil
}

// Enforcer exposes the enforcer for use in route registration.
func (h *Handler) Enforcer() *casbin.Enforcer { return h.enforcer }
