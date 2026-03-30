package handlers

import (
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services"
)

type Handler struct {
	// for cross-service api calls, the orchestration handler app service access
	appSvc       *services.AppService
	ETL          *etl.Handler
	sessions     *sessionStore
	Optimization *optimization.Handler
}

func NewHandler(appSvc *services.AppService, cfg *appconfig.Config, db *database.Database) (*Handler, error) {
	sessionStore, err := newSessionStore(cfg, db)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		appSvc:   appSvc,
		sessions: sessionStore,
		ETL:      etl.NewHandler(appSvc.ETL()),
	}
	if opt := appSvc.Optimization(); opt != nil {
		h.Optimization = optimization.NewHandler(opt)
	}

	return h, nil
}
