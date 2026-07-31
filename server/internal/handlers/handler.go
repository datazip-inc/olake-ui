package handlers

import (
	"os"
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
)

// acts as the orchestration layer for: ETL & Optimization handlers
type Handler struct {
	// for cross-service api calls, the orchestration handler has app service access
	appSvc       *services.AppService
	ETL          *etl.Handler
	Optimization *optimization.Handler
	sessions     *sessionStore
}

func NewHandler(appSvc *services.AppService, cfg *appconfig.Config, db *database.Database) *Handler {
	sessionStore := newSessionStore(cfg, db)

	handler := &Handler{
		appSvc:   appSvc,
		ETL:      etl.NewHandler(appSvc.ETL()),
		sessions: sessionStore,
	}

	if opt := appSvc.Optimization(); opt != nil {
		handler.Optimization = optimization.NewHandler(opt)
		KubernetesServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
		telemetry.TrackInstalledFusion(KubernetesServiceHost)
	}

	return handler
}
