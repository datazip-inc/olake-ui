package etl

import (
	"github.com/gin-gonic/gin"

	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// encapsulates ETL-specific request handling and business logic.
type Handler struct {
	etl *services.Service
}

// NewHandler initializes the ETL handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	return &Handler{etl: s}
}

// Metrics serves GET /metrics: refresh the snapshot (the previous one is served on
// error), then write the Prometheus exposition — same delegation shape as ServeSwagger.
func (h *Handler) Metrics(c *gin.Context) {
	exposition, err := h.etl.RefreshMetrics(c.Request.Context())
	if err != nil {
		logger.Errorf("metrics refresh failed, serving previous snapshot: %s", err)
	}
	exposition.ServeHTTP(c.Writer, c.Request)
}
