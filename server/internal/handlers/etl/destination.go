package etl

import (
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @Summary List all destinations
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

// @Summary Get destination details
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

// @Summary Test destination connection
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
	if err := dto.UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
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

// @Summary Get available destination versions
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

// @Summary Get destination UI spec
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
	if err := dto.UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
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
