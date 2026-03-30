package etl

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
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
func (h *Handler) ListDestinations(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get all destinations initiated project_id[%s]", projectID)
	items, err := h.etl.ListDestinations(c.Request.Context(), projectID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to get destinations: %s", err), err)
		return
	}
	successResponse(c, "destinations listed successfully", items)
}

// @Summary Get destination details
// @Tags Destinations
// @Description Retrieve details of a specific destination identified by its unique ID.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Success 200 {object} dto.JSONResponse{data=dto.DestinationDataItem}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "destination not found"
// @Failure 500 {object} dto.Error500Response "failed to get destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [get]
func (h *Handler) GetDestination(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	destinationID, err := getIDParam(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get destination initiated project_id[%s] destination_id[%d]", projectID, destinationID)
	destination, err := h.etl.GetDestination(c.Request.Context(), projectID, destinationID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrDestinationNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(c, status, fmt.Sprintf("failed to get destination: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination '%d' retrieved successfully", destinationID), destination)
}

// @Summary Create a new destination
// @Tags Destinations
// @Description Create a new destination for a project
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.CreateDestinationRequest true "destination data"
// @Success 200 {object} dto.JSONResponse{data=dto.CreateDestinationRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to create destination"
// @Router /api/v1/project/{projectid}/destinations [post]
func (h *Handler) CreateDestination(c *gin.Context) {
	userID := getCurrentUserID(c, h.sessions)
	if userID == nil {
		errorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.CreateDestinationRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := dto.ValidateDestinationType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Create destination initiated project_id[%s] destination_type[%s] destination_name[%s] user_id[%v]",
		projectID, req.Type, req.Name, userID)
	if err := h.etl.CreateDestination(c.Request.Context(), &req, projectID, userID); err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to create destination: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination %s created successfully", req.Name), req)
}

// @Summary Update a destination
// @Tags Destinations
// @Description Update the configuration details of an existing destination.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Param   body          body    dto.UpdateDestinationRequest true "destination data"
// @Success 200 {object} dto.JSONResponse{data=dto.UpdateDestinationRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "destination not found"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to update destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [put]
func (h *Handler) UpdateDestination(c *gin.Context) {
	userID := getCurrentUserID(c, h.sessions)
	if userID == nil {
		errorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}
	id, err := getIDParam(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.UpdateDestinationRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := dto.ValidateDestinationType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Update destination initiated project_id[%s], destination_id[%d], destination_type[%s], user_id[%v]",
		projectID, id, req.Type, userID)

	if err := h.etl.UpdateDestination(c.Request.Context(), id, projectID, &req, userID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrDestinationNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(c, status, fmt.Sprintf("failed to update destination: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination %s updated successfully", req.Name), req)
}

// @Summary Delete a destination
// @Tags Destinations
// @Description Permanently delete a specified destination.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "destination id"
// @Success 200 {object} dto.JSONResponse{data=dto.DeleteDestinationResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "destination not found"
// @Failure 500 {object} dto.Error500Response "failed to delete destination"
// @Router /api/v1/project/{projectid}/destinations/{id} [delete]
func (h *Handler) DeleteDestination(c *gin.Context) {
	id, err := getIDParam(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Delete destination initiated destination_id[%d]", id)
	resp, err := h.etl.DeleteDestination(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrDestinationNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(c, status, fmt.Sprintf("failed to delete destination: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination %s deleted successfully", resp.Name), resp)
}

// @Summary Test destination connection
// @Tags Destinations
// @Description Validate the connection to a destination using the provided configuration details.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.DestinationTestConnectionRequest true "test connection data"
// @Success 200 {object} dto.JSONResponse{data=dto.TestConnectionResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Router /api/v1/project/{projectid}/destinations/test [post]
func (h *Handler) TestDestinationConnection(c *gin.Context) {
	// need to remove sourceVersion from request
	var req dto.DestinationTestConnectionRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Test destination connection initiated destination_type[%s] destination_version[%s]", req.Type, req.Version)

	result, logs, err := h.etl.TestDestinationConnection(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to verify driver credentials: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination %s connection tested successfully", req.Type), dto.TestConnectionResponse{
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
func (h *Handler) GetDestinationVersions(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	destType := c.Query("type")
	if err := dto.ValidateDestinationType(destType); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get destination versions initiated project_id[%s] destination_type[%s]", projectID, destType)

	versions, err := h.etl.GetDestinationVersions(c.Request.Context(), destType)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to get destination versions: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination %s versions fetched successfully", destType), versions)
}

// @Summary Get destination UI spec
// @Tags Destinations
// @Description Retrieve the UI spec for a specific destination type/version.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.SpecRequest true "spec request data"
// @Success 200 {object} dto.JSONResponse{data=dto.SpecResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to get spec"
// @Router /api/v1/project/{projectid}/destinations/spec [post]
func (h *Handler) GetDestinationSpec(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.SpecRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get destination spec initiated project_id[%s] destination_type[%s] destination_version[%s]",
		projectID, req.Type, req.Version)
	resp, err := h.etl.GetDestinationSpec(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to get destination spec: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("destination %s spec fetched successfully", req.Type), resp)
}
