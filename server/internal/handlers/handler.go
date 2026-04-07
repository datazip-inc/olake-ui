package handlers

import (
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services"
)

// acts as the orchestration layer for: ETL & Optimization handlers
type Handler struct {
	// for cross-service api calls, the orchestration handler has app service access
	appSvc       *services.AppService
	ETL          *etl.Handler
	Optimization *optimization.Handler
	sessions     *sessionStore
}

func NewHandler(appSvc *services.AppService, cfg *appconfig.Config, db *database.Database) (*Handler, error) {
	sessionStore := newSessionStore(cfg, db)

	h := &Handler{
		appSvc:   appSvc,
		ETL:      etl.NewHandler(appSvc.ETL()),
		sessions: sessionStore,
	}

	if opt := appSvc.Optimization(); opt != nil {
		h.Optimization = optimization.NewHandler(opt)
	}

	return h, nil
}
