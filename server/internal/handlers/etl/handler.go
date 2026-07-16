package etl

import (
	"net/http"

	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

// encapsulates ETL-specific request handling and business logic.
type Handler struct {
	etl *services.Service
}

// NewHandler initializes the ETL handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	return &Handler{etl: s}
}

// MetricsHandler exposes the ETL service's Prometheus /metrics handler.
func (h *Handler) MetricsHandler() http.Handler {
	return h.etl.MetricsHandler()
}
