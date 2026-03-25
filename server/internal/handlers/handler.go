package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimisation"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/utils"
)

// acts as the orchestration layer that composes domain-specific handlers.
type Handler struct {
	web.Controller
	ETL          *etl.Handler
	Optimisation *optimisation.Handler
}

// NewHandler creates the orchestration handler by composing domain handlers.
// It conditionally initializes optimisation support based on service availability.
func NewHandler(appSvc *services.AppService) *Handler {
	h := &Handler{
		ETL: etl.NewHandler(appSvc.ETL()),
	}
	if appSvc.Optimisation() != nil {
		h.Optimisation = optimisation.NewHandler(appSvc.Optimisation())
	}

	return h
}

// GetoptimisationStatus reports whether optimisation features are enabled.
func (h *Handler) GetoptimisationStatus() {
	response := map[string]interface{}{
		"enabled": h.Optimisation != nil,
	}

	utils.SuccessResponse(&h.Controller, "optimisation status retrieved successfully", response)
}
