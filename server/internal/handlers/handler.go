package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// acts as the orchestration layer that composes domain-specific handlers.
type Handler struct {
	web.Controller
	ETL          *etl.Handler
	Optimization *optimization.Handler
}

// NewHandler creates the orchestration handler by composing domain handlers.
// It conditionally initializes optimization support based on service availability.
func NewHandler(appSvc *services.AppService) *Handler {
	h := &Handler{
		ETL: etl.NewHandler(appSvc.ETL()),
	}
	if appSvc.Optimization() != nil {
		h.Optimization = optimization.NewHandler(appSvc.Optimization())
	}

	return h
}

// GetoptimizationStatus reports whether optimization features are enabled.
func (h *Handler) GetoptimizationStatus() {
	response := map[string]interface{}{
		"enabled": h.Optimization != nil,
	}

	utils.SuccessResponse(&h.Controller, "optimization status retrieved successfully", response)
}
