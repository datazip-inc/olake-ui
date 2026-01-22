package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @Title GetReleaseUpdates
// @Tags Platform
// @Description Retrieve the latest platform release updates and metadata.
// @Param   limit         query   int     false   "limit the number of releases returned"
// @Success 200 {object} dto.JSONResponse{data=dto.ReleasesResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to fetch release metadata"
// @Router /api/v1/platform/releases [get]
func (h *Handler) GetReleaseUpdates() {
	limitStr := h.Ctx.Input.Query("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logger.Debugf("Get release metadata initiated limit[%d]", limit)

	response, err := h.etl.GetAllReleasesResponse(h.Ctx.Request.Context(), limit)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to fetch release metadata: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, "release metadata fetched successfully", response)
}
