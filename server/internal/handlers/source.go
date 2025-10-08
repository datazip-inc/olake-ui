package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/utils"
)

type SourceHandler struct {
	web.Controller
	sourceService *services.SourceService
}

func (c *SourceHandler) Prepare() {
	var err error
	c.sourceService, err = services.NewSourceService()
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to initialize source service", err)
		return
	}
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	projectID := c.Ctx.Input.Param(":projectid")
	sources, err := c.sourceService.GetAllSources(c.Ctx.Request.Context(), projectID)
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
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}


	userID := GetUserIDFromSession(&c.Controller)
	if err := c.sourceService.CreateSource(c.Ctx.Request.Context(), req, projectID, userID); err != nil {
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
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}


	userID := GetUserIDFromSession(&c.Controller)
	if err := c.sourceService.UpdateSource(context.Background(), id, req, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update source", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id := GetIDFromPath(&c.Controller)

	resp, err := c.sourceService.DeleteSource(c.Ctx.Request.Context(), id)
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
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	result, err := c.sourceService.TestConnection(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to test connection", err)
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /sources/streams [post]
func (c *SourceHandler) GetSourceCatalog() {
	var req dto.StreamsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	catalog, err := c.sourceService.GetSourceCatalog(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source catalog", err)
		return
	}
	utils.SuccessResponse(&c.Controller, catalog)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	id := GetIDFromPath(&c.Controller)
	jobs, err := c.sourceService.GetSourceJobs(c.Ctx.Request.Context(), id)
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
	sourceType := c.GetString("type")
	versions, err := c.sourceService.GetSourceVersions(context.Background(), sourceType)
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
// @router /project/:projectid/sources/spec [post]
func (c *SourceHandler) GetProjectSourceSpec() {
	var req dto.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	specOutput, err := c.sourceService.GetSourceSpec(c.Ctx.Request.Context(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source spec", err)
		return
	}

	utils.SuccessResponse(&c.Controller, resp)
}
