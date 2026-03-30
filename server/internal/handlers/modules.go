package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httputil"
)

func (h *Handler) GetOptimizationStatus(c *gin.Context) {
	httputil.SuccessResponse(c, "optimization status retrieved successfully", map[string]interface{}{
		"enabled": h.appSvc.Optimization() != nil,
	})
}
