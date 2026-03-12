package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @Summary List all sources
// @Tags Sources
// @Description Retrieve a list of all configured sources within a specific project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Success 200 {object} dto.JSONResponse{data=[]dto.SourceDataItem}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to retrieve sources"
// @Router /api/v1/project/{projectid}/sources [get]
func (h *Handler) ListSources(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get all sources initiated project_id[%s]", projectID)
	sources, err := h.etl.ListSources(c.Request.Context(), projectID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve sources: %s", err), err)
		return
	}
	successResponse(c, "sources listed successfully", sources)
}

// @Summary Get source details
// @Tags Sources
// @Description Retrieve details of a specific source identified by its unique ID.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "source id"
// @Success 200 {object} dto.JSONResponse{data=dto.SourceDataItem}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get source"
// @Router /api/v1/project/{projectid}/sources/{id} [get]
func (h *Handler) GetSource(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	sourceID, err := getIDParam(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get source initiated project_id[%s] source_id[%d]", projectID, sourceID)
	source, err := h.etl.GetSource(c.Request.Context(), projectID, sourceID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrSourceNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(c, status, fmt.Sprintf("failed to get source: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source '%d' retrieved successfully", sourceID), source)
}

// @Summary Create a new source
// @Tags Sources
// @Description Create a new source within a project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.CreateSourceRequest true "source data"
// @Success 200 {object} dto.JSONResponse{data=dto.CreateSourceRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to create source"
// @Router /api/v1/project/{projectid}/sources [post]
func (h *Handler) CreateSource(c *gin.Context) {
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
	var req dto.CreateSourceRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := dto.ValidateSourceType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Create source initiated project_id[%s] source_type[%s] source_name[%s] user_id[%v]", projectID, req.Type, req.Name, userID)
	if err := h.etl.CreateSource(c.Request.Context(), &req, projectID, userID); err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to create source: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s created successfully", req.Name), req)
}

// @Summary Update a source
// @Tags Sources
// @Description Update the configuration details of an existing source.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "source id"
// @Param   body          body    dto.UpdateSourceRequest true "source data"
// @Success 200 {object} dto.JSONResponse{data=dto.UpdateSourceRequest}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "source not found"
// @Failure 500 {object} dto.Error500Response "failed to update source"
// @Router /api/v1/project/{projectid}/sources/{id} [put]
func (h *Handler) UpdateSource(c *gin.Context) {
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
	id, err := getIDParam(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.UpdateSourceRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := dto.ValidateSourceType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Update source initiated project_id[%s] source_id[%d] source_type[%s] user_id[%v]", projectID, id, req.Type, userID)
	if err := h.etl.UpdateSource(c.Request.Context(), projectID, id, &req, userID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrSourceNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(c, status, fmt.Sprintf("failed to update source: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s updated successfully", req.Name), req)
}

// @Summary Delete a source
// @Tags Sources
// @Description Permanently delete a specified source.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "source id"
// @Success 200 {object} dto.JSONResponse{data=dto.DeleteSourceResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "source not found"
// @Failure 500 {object} dto.Error500Response "failed to delete source"
// @Router /api/v1/project/{projectid}/sources/{id} [delete]
func (h *Handler) DeleteSource(c *gin.Context) {
	id, err := getIDParam(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Delete source initiated source_id[%d]", id)
	resp, err := h.etl.DeleteSource(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrSourceNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(c, status, fmt.Sprintf("failed to delete source: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s deleted successfully", resp.Name), resp)
}

// @Summary Test source connection
// @Tags Sources
// @Description Validate the connection to a source using the provided configuration details.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.SourceTestConnectionRequest true "test connection data"
// @Success 200 {object} dto.JSONResponse{data=dto.TestConnectionResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to test connection"
// @Router /api/v1/project/{projectid}/sources/test [post]
func (h *Handler) TestSourceConnection(c *gin.Context) {
	var req dto.SourceTestConnectionRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := dto.ValidateSourceType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Infof("Test source connection initiated source_type[%s] source_version[%s]", req.Type, req.Version)
	result, logs, err := h.etl.TestSourceConnection(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to verify credentials: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s connection tested successfully", req.Type), dto.TestConnectionResponse{
		ConnectionResult: result,
		Logs:             logs,
	})
}

// @Summary Get source stream catalog
// @Tags Sources
// @Description Discover and list available data streams from a source.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.StreamsRequest true "streams request data"
// @Success 200 {object} dto.JSONResponse{data=object}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get source catalog"
// @Router /api/v1/project/{projectid}/sources/streams [post]
func (h *Handler) GetSourceCatalog(c *gin.Context) {
	var req dto.StreamsRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, statusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := dto.ValidateSourceType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get source catalog initiated source_type[%s] source_version[%s] job_id[%d]", req.Type, req.Version, req.JobID)
	catalog, err := h.etl.GetSourceCatalog(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to get source streams: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s catalog fetched successfully", req.Type), catalog)
}

// @Summary Get available source versions
// @Tags Sources
// @Description Retrieve the list of available versions for a specific source connector type.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   type          query   string  true    "source type"
// @Success 200 {object} dto.JSONResponse{data=dto.VersionsResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get versions"
// @Router /api/v1/project/{projectid}/sources/versions [get]
func (h *Handler) GetSourceVersions(c *gin.Context) {
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	sourceType := c.Query("type")
	if sourceType == "" {
		errorResponse(c, http.StatusBadRequest, "failed to get source versions: source type is required", fmt.Errorf("source type is required"))
		return
	}
	logger.Debugf("Get source versions initiated project_id[%s] source_type[%s]", projectID, sourceType)
	versions, err := h.etl.GetSourceVersions(c.Request.Context(), sourceType)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to get source versions: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s versions fetched successfully", sourceType), versions)
}

// @Summary Get source UI spec
// @Tags Sources
// @Description Retrieve the UI spec for a specific source type/version.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.SpecRequest true "spec request data"
// @Success 200 {object} dto.JSONResponse{data=dto.SpecResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get spec"
// @Router /api/v1/project/{projectid}/sources/spec [post]
func (h *Handler) GetSourceSpec(c *gin.Context) {
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
	if err := dto.ValidateSourceType(req.Type); err != nil {
		errorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get source spec initiated project_id[%s] source_type[%s] source_version[%s]", projectID, req.Type, req.Version)
	resp, err := h.etl.GetSourceSpec(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to get source spec: %s", err), err)
		return
	}
	successResponse(c, fmt.Sprintf("source %s spec fetched successfully", req.Type), resp)
}
