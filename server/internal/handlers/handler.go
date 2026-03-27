package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// acts as the orchestration layer for: ETL & Optimization handlers
type Handler struct {
	web.Controller
	// for cross-service api calls, the orchestration handler
	// has app service access
	appSvc       *services.AppService
	ETL          *etl.Handler
	Optimization *optimization.Handler
}

var app *services.AppService

// domain-specific handler
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

// Prepare runs before each action; Beego constructs a fresh controller per request,
// so we assign the shared AppService here to avoid nil dereferences.
func (h *Handler) Prepare() {
	h.appSvc = app
}

func (h *Handler) GetoptimizationStatus() {
	response := map[string]interface{}{
		"enabled": h.appSvc.Optimization() != nil,
	}

	utils.SuccessResponse(&h.Controller, "optimization status retrieved successfully", response)
}
