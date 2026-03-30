package handlers

import (
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

var app *services.AppService

func NewHandler(appSvc *services.AppService) *Handler {
	app = appSvc
	h := &Handler{
		appSvc: appSvc,
		ETL:    etl.NewHandler(appSvc.ETL()),
	}
	if appSvc.Optimization() != nil {
		h.Optimization = optimization.NewHandler(appSvc.Optimization())
	}

	return h
}

func (h *Handler) GetoptimizationStatus() {
	response := map[string]interface{}{
		"enabled": h.appSvc.Optimization() != nil,
	}

	successResponse(c, "optimization status retrieved successfully", response)
}
