package handlers

import (
	"errors"
	"net/http"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/utils"
)

// @router /project/:projectid/sources [get]
func (h *Handler) ListSources() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Get all sources initiated project_id[%s]", projectID)

	sources, err := h.svc.GetAllSources(h.Ctx.Request.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to retrieve sources", err)
		return
	}
	utils.SuccessResponse(&h.Controller, sources)
}

// @router /project/:projectid/sources [post]
func (h *Handler) CreateSource() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}

	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	var req dto.CreateSourceRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := dto.ValidateSourceType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Create source initiated project_id[%s] source_type[%s] source_name[%s] user_id[%v]",
		projectID, req.Type, req.Name, userID)

	if err := h.svc.CreateSource(h.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to create source", err)
		return
	}

	utils.SuccessResponse(&h.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (h *Handler) UpdateSource() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	var req dto.UpdateSourceRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := dto.ValidateSourceType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Update source initiated project_id[%s] source_id[%d] source_type[%s] user_id[%v]",
		projectID, id, req.Type, userID)

	if err := h.svc.UpdateSource(h.Ctx.Request.Context(), projectID, id, &req, userID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrSourceNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(&h.Controller, status, "Failed to update source", err)
		return
	}

	utils.SuccessResponse(&h.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (h *Handler) DeleteSource() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Delete source initiated source_id[%d]", id)

	resp, err := h.svc.DeleteSource(h.Ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			utils.ErrorResponse(&h.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to delete source", err)
		}
		return
	}
	utils.SuccessResponse(&h.Controller, resp)
}

// @router /project/:projectid/sources/test [post]
func (h *Handler) TestSourceConnection() {
	var req dto.SourceTestConnectionRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := dto.ValidateSourceType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Infof("Test source connection initiated source_type[%s] source_version[%s]", req.Type, req.Version)

	result, logs, err := h.svc.SourceTestConnection(h.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to test connection", err)
		return
	}

	utils.SuccessResponse(&h.Controller, dto.TestConnectionResponse{
		ConnectionResult: result,
		Logs:             logs,
	})
}

// @router /sources/streams [post]
func (h *Handler) GetSourceCatalog() {
	var req dto.StreamsRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := dto.ValidateSourceType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Get source catalog initiated source_type[%s] source_version[%s] job_id[%d]",
		req.Type, req.Version, req.JobID)

	catalog, err := h.svc.GetSourceCatalog(h.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get source catalog", err)
		return
	}
	utils.SuccessResponse(&h.Controller, catalog)
}

// @router /sources/:id/jobs [get]
func (h *Handler) GetSourceJobs() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Get source jobs initiated source_id[%d]", id)

	jobs, err := h.svc.GetSourceJobs(h.Ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			utils.ErrorResponse(&h.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get source jobs", err)
		}
		return
	}
	utils.SuccessResponse(&h.Controller, map[string]interface{}{"jobs": jobs})
}

// @router /project/:projectid/sources/versions [get]
func (h *Handler) GetSourceVersions() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	sourceType := h.GetString("type")
	logger.Debugf("Get source versions initiated project_id[%s] source_type[%s]", projectID, sourceType)

	versions, err := h.svc.GetSourceVersions(h.Ctx.Request.Context(), sourceType)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "source type is required" {
			status = http.StatusBadRequest
		}
		utils.ErrorResponse(&h.Controller, status, "Failed to get source versions", err)
		return
	}
	utils.SuccessResponse(&h.Controller, versions)
}

// @router /project/:projectid/sources/spec [post]
func (h *Handler) GetProjectSourceSpec() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	var req dto.SpecRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := dto.ValidateSourceType(req.Type); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Get source spec initiated project_id[%s] source_type[%s] source_version[%s]",
		projectID, req.Type, req.Version)

	resp, err := h.svc.GetSourceSpec(h.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get source spec", err)
		return
	}

	utils.SuccessResponse(&h.Controller, resp)
}
