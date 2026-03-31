package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httpx"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// @Summary Create a new destination in ETL & catalog in Optimization (if enabled)
// @Tags Destinations
// @Description Create a new destination for a project
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.CreateDestinationRequest true "destination data"
// @Success 200 {object} dto.JSONResponse{data=dto.CreateDestinationRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to create destination"
// @Router /api/v1/project/{projectid}/destinations [post]
func (h *Handler) CreateDestinationAndCatalog(c *gin.Context) {
	userID := httpx.GetCurrentUserID(c)
	if userID == nil {
		httpx.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	projectID, err := httpx.GetProjectID(c)
	if err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.CreateDestinationRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := dto.ValidateDestinationType(req.Type); err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := h.appSvc.CreateDestinationWithCatalog(c.Request.Context(), projectID, &req, userID); err != nil {
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to create destination & catalog: %s", err), err)
		return
	}

	httpx.SuccessResponse(c, fmt.Sprintf("destination %s and catalog created successfully", req.Name), req)
}

// @Summary Update a destination in ETL & catalog in Optimization (if enabled)
// @Tags Destinations
// @Description Update the configuration details of an existing destination.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Param   body          body    dto.UpdateDestinationRequest true "destination data"
// @Success 200 {object} dto.JSONResponse{data=dto.UpdateDestinationRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to update destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [put]
func (h *Handler) UpdateDestinationAndCatalog(c *gin.Context) {
	userID := httpx.GetCurrentUserID(c)
	if userID == nil {
		httpx.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	id, err := httpx.GetIDParam(c, "id")
	if err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	projectID, err := httpx.GetProjectID(c)
	if err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpdateDestinationRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := dto.ValidateDestinationType(req.Type); err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := h.appSvc.UpdateDestinationWithCatalog(c.Request.Context(), id, projectID, &req, userID); err != nil {
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to update destination: %s", err), err)
		return
	}

	httpx.SuccessResponse(c, fmt.Sprintf("destination %s updated successfully", req.Name), req)
}

// @Summary Delete a destination in ETL & catalog in Optimization (if enabled)
// @Tags Destinations
// @Description Permanently delete a specified destination.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Success 200 {object} dto.JSONResponse{data=dto.DeleteDestinationResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to delete destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [delete]
func (h *Handler) DeleteDestinationAndCatalog(c *gin.Context) {
	id, err := httpx.GetIDParam(c, "id")
	if err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	resp, err := h.appSvc.DeleteDestinationWithCatalog(c.Request.Context(), id)
	if err != nil {
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete destination: %s", err), err)
		return
	}

	httpx.SuccessResponse(c, fmt.Sprintf("destination %s deleted successfully as well as catalog", resp.Name), resp)
}
