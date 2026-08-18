package etl

import (
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

// Metrics serves GET /metrics: refresh the snapshot (the previous one is served on error), then write the Prometheus exposition
func (h *Handler) Metrics(c *gin.Context) {
	exposition, err := h.etl.RefreshMetrics(c.Request.Context())
	if err != nil {
		logger.Errorf("metrics refresh failed, serving previous snapshot: %s", err)
	}
	exposition.ServeHTTP(c.Writer, c.Request)
}
