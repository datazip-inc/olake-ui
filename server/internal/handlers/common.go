package handlers

import (
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
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
func (h *Handler) CreateDestinationAndCatalog() {
	userID := etl.GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	projectID, err := etl.GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.CreateDestinationRequest
	if err := dto.UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := dto.ValidateDestinationType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Create destination initiated project_id[%s] destination_type[%s] destination_name[%s] user_id[%v]",
		projectID, req.Type, req.Name, userID)

	result, err := h.appSvc.CreateDestinationWithCatalog(h.Ctx.Request.Context(), projectID, &req, userID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create destination: %s", err), err)
		return
	}
	if result.CatalogErr != nil {
		utils.ErrorResponse(&h.Controller, http.StatusPartialContent, "destination created but catalog creation failed", result.CatalogErr)
		return
	}

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s created successfully", req.Name), req)
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
func (h *Handler) UpdateDestinationAndCatalog() {
	userID := etl.GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	id, err := etl.GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	projectID, err := etl.GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpdateDestinationRequest
	if err := dto.UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := dto.ValidateDestinationType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Update destination initiated project_id[%s] destination_id[%d] destination_type[%s] user_id[%v]",
		projectID, id, req.Type, userID)

	result, err := h.appSvc.UpdateDestinationWithCatalog(h.Ctx.Request.Context(), id, projectID, &req, userID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update destination: %s", err), err)
		return
	}
	if result.CatalogErr != nil {
		utils.ErrorResponse(&h.Controller, http.StatusPartialContent, "destination updated but catalog updation failed", result.CatalogErr)
		return
	}

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s updated successfully", req.Name), req)
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
func (h *Handler) DeleteDestinationAndCatalog() {
	id, err := etl.GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Delete destination initiated destination_id[%d]", id)

	resp, result, err := h.appSvc.DeleteDestinationWithCatalog(h.Ctx.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to delete destination: %s", err), err)
		return
	}
	if result.CatalogErr != nil {
		utils.ErrorResponse(&h.Controller, http.StatusPartialContent, "destination deleted but catalog deletion failed", result.CatalogErr)
		return
	}

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s deleted successfully as well as catalog", resp.Name), resp)
}
