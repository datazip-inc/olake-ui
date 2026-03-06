package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/datazip-inc/olake-ui/server/utils/logger"

)

// @Summary Get release updates
// @Tags Platform
// @Description Retrieve the latest platform release updates and metadata.
// @Param   limit         query   int     false   "limit the number of releases returned"
// @Success 200 {object} dto.JSONResponse{data=dto.ReleasesResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to fetch release metadata"
// @Router /api/v1/platform/releases [get]
func (h *GinHandler) getReleaseUpdates(c *gin.Context) {
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	logger.Debugf("Get release updates initiated limit[%d]", limit)
	response, err := h.etl.GetAllReleasesResponse(c.Request.Context(), limit)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to fetch release metadata: %s", err), err)
		return
	}
	successResponse(c, "release metadata fetched successfully", response)
}
