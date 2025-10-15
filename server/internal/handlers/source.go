package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/utils"
)

type SourceHandler struct {
	web.Controller
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	projectID := c.Ctx.Input.Param(":projectid")
	logger.Info("Get all sources initiated - project_id=%s", projectID)

	sources, err := SourceSvc().GetAllSources(c.Ctx.Request.Context(), projectID)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources", err)
		return
	}
	utils.SuccessResponse(&c.Controller, sources)
}

// @router /project/:projectid/sources [post]
func (c *SourceHandler) CreateSource() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.CreateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Create source initiated - project_id=%s source_type=%s source_name=%s user_id=%v",
		projectID, req.Type, req.Name, userID)

	if err := SourceSvc().CreateSource(c.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create source", err)
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	id := GetIDFromPath(&c.Controller)
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.UpdateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Update source initiated - project_id=%s source_id=%d source_type=%s user_id=%v",
		projectID, id, req.Type, userID)

	if err := SourceSvc().UpdateSource(c.Ctx.Request.Context(), projectID, id, &req, userID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrSourceNotFound) {
			status = http.StatusNotFound
		}
		respondWithError(&c.Controller, status, "Failed to update source", err)
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id := GetIDFromPath(&c.Controller)
	logger.Info("Delete source initiated - source_id=%d", id)

	resp, err := SourceSvc().DeleteSource(c.Ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			respondWithError(&c.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete source", err)
		}
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req dto.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Test source connection initiated - source_type=%s source_version=%s", req.Type, req.Version)

	result, logs, err := SourceSvc().TestConnection(c.Ctx.Request.Context(), &req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to test connection", err)
		return
	}

	utils.SuccessResponse(&c.Controller, dto.TestConnectionResponse{
		ConnectionResult: result,
		Logs:             logs,
	})
}

// @router /sources/streams [post]
func (c *SourceHandler) GetSourceCatalog() {
	var req dto.StreamsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Get source catalog initiated - source_type=%s source_version=%s job_id=%d",
		req.Type, req.Version, req.JobID)

	catalog, err := SourceSvc().GetSourceCatalog(c.Ctx.Request.Context(), &req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source catalog", err)
		return
	}
	utils.SuccessResponse(&c.Controller, catalog)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	id := GetIDFromPath(&c.Controller)
	logger.Info("Get source jobs initiated - source_id=%d", id)

	jobs, err := SourceSvc().GetSourceJobs(c.Ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			respondWithError(&c.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source jobs", err)
		}
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"jobs": jobs})
}

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	projectID := c.Ctx.Input.Param(":projectid")
	sourceType := c.GetString("type")
	logger.Info("Get source versions initiated - project_id=%s source_type=%s", projectID, sourceType)

	versions, err := SourceSvc().GetSourceVersions(c.Ctx.Request.Context(), sourceType)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "source type is required" {
			status = http.StatusBadRequest
		}
		respondWithError(&c.Controller, status, "Failed to get source versions", err)
		return
	}
	utils.SuccessResponse(&c.Controller, versions)
}

// @router /project/:projectid/sources/spec [post]
func (c *SourceHandler) GetProjectSourceSpec() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Get source spec initiated - project_id=%s source_type=%s source_version=%s",
		projectID, req.Type, req.Version)

	resp, err := SourceSvc().GetSourceSpec(c.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, resp)
}
