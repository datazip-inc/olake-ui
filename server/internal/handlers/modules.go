package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

func (h *Handler) GetOptimizationStatus(c *gin.Context) {
	utils.SuccessResponse(c, "optimization status retrieved successfully", map[string]interface{}{
		"enabled": h.appSvc.Optimization() != nil,
	})
}
