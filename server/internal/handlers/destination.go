package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @router /project/:projectid/destinations [get]
func (h *Handler) ListDestinations() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	items, err := h.etl.ListDestinations(h.Ctx.Request.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get destinations: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, items)
}

// @router /project/:projectid/destinations [post]
func (h *Handler) CreateDestination() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}

	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.CreateDestinationRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := dto.ValidateDestinationType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Create destination initiated project_id[%s] destination_type[%s] destination_name[%s] user_id[%v]",
		projectID, req.Type, req.Name, userID)

	if err := h.etl.CreateDestination(h.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create destination: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (h *Handler) UpdateDestination() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpdateDestinationRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := dto.ValidateDestinationType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Update destination initiated project_id[%s], destination_id[%d], destination_type[%s], user_id[%v]",
		projectID, id, req.Type, userID)

	if err := h.etl.UpdateDestination(h.Ctx.Request.Context(), id, projectID, &req, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update destination: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (h *Handler) DeleteDestination() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Delete destination initiated destination_id[%d]", id)

	resp, err := h.etl.DeleteDestination(h.Ctx.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to delete destination: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, resp)
}

// @router /project/:projectid/destinations/test [post]
func (h *Handler) TestDestinationConnection() {
	var req dto.DestinationTestConnectionRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Test destination connection initiated destination_type[%s] destination_version[%s]", req.Type, req.Version)

	result, logs, err := h.etl.TestConnection(h.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to verify driver credentials: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, dto.TestConnectionResponse{
		ConnectionResult: result,
		Logs:             logs,
	})
}

// @router /destinations/:id/jobs [get]
func (h *Handler) GetDestinationJobs() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get destination jobs initiated destination_id[%d]", id)

	jobs, err := h.etl.GetDestinationJobs(h.Ctx.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get jobs related to destination: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, map[string]interface{}{"jobs": jobs})
}

// @router /project/:projectid/destinations/versions [get]
func (h *Handler) GetDestinationVersions() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	destType := h.GetString("type")
	if err := dto.ValidateDestinationType(destType); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get destination versions initiated project_id[%s] destination_type[%s]", projectID, destType)

	versions, err := h.etl.GetDestinationVersions(h.Ctx.Request.Context(), destType)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to get destination versions: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, versions)
}

// @router /project/:projectid/destinations/spec [post]
func (h *Handler) GetDestinationSpec() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.SpecRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get destination spec initiated project_id[%s] destination_type[%s] destination_version[%s]",
		projectID, req.Type, req.Version)

	resp, err := h.etl.GetDestinationSpec(h.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get destination spec: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, resp)
}
