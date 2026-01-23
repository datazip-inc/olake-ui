package handlers

import (
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @Title ListDestinations
// @Tags Destinations
// @Description Retrieve a list of all configured destinations within a specific project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Success 200 {object} dto.JSONResponse{data=[]dto.DestinationDataItem}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get destinations"
// @Router /api/v1/project/{projectid}/destinations [get]
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
	utils.SuccessResponse(&h.Controller, "Destinations listed successfully", items)
}

// @Title GetDestination
// @Tags Destinations
// @Description Retrieve details of a specific destination identified by its unique ID.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Success 200 {object} dto.JSONResponse{data=dto.DestinationDataItem}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [get]
func (h *Handler) GetDestination() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	destinationID, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get destination initiated project_id[%s] destination_id[%d]", projectID, destinationID)

	destination, err := h.etl.GetDestination(h.Ctx.Request.Context(), projectID, destinationID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get destination: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination '%d' retrieved successfully", destinationID), destination)
}

// @Title CreateDestination
// @Tags Destinations
// @Description Create a new destination for a project
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.CreateDestinationRequest true "destination data"
// @Success 200 {object} dto.JSONResponse{data=dto.CreateDestinationRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to create destination"
// @Router /api/v1/project/{projectid}/destinations [post]
func (h *Handler) CreateDestination() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
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

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s created successfully", req.Name), req)
}

// @Title UpdateDestination
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
func (h *Handler) UpdateDestination() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
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
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s updated successfully", req.Name), req)
}

// @Title DeleteDestination
// @Tags Destinations
// @Description Permanently delete a specified destination.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Success 200 {object} dto.JSONResponse{data=dto.DeleteDestinationResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to delete destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [delete]
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

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s deleted successfully", resp.Name), resp)
}

// @Title TestDestinationConnection
// @Tags Destinations
// @Description Validate the connection to a destination using the provided configuration details.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.DestinationTestConnectionRequest true "test connection data"
// @Success 200 {object} dto.JSONResponse{data=dto.TestConnectionResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to test connection"
// @Router /api/v1/project/{projectid}/destinations/test [post]
func (h *Handler) TestDestinationConnection() {
	// need to remove sourceVersion from request
	var req dto.DestinationTestConnectionRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Test destination connection initiated destination_type[%s] destination_version[%s]", req.Type, req.Version)

	result, logs, err := h.etl.TestDestinationConnection(h.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to verify driver credentials: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s connection tested successfully", req.Type), dto.TestConnectionResponse{
		ConnectionResult: result,
		Logs:             logs,
	})
}

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
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %d jobs fetched successfully", id), map[string]interface{}{"jobs": jobs})
}

// @Title GetDestinationVersions
// @Tags Destinations
// @Description Retrieve the list of available versions for a specific destination connector type.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   type          query   string  true    "destination type"
// @Success 200 {object} dto.JSONResponse{data=dto.VersionsResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get versions"
// @Router /api/v1/project/{projectid}/destinations/versions [get]
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
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s versions fetched successfully", destType), versions)
}

// @Title GetDestinationSpec
// @Tags Destinations
// @Description Retrieve the UI spec for a specific destination type/version.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.SpecRequest true "spec request data"
// @Success 200 {object} dto.JSONResponse{data=dto.SpecResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get spec"
// @Router /api/v1/project/{projectid}/destinations/spec [post]
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
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("destination %s spec fetched successfully", req.Type), resp)
}
