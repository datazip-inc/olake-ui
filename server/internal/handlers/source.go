package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	c.sourceService, err = services.NewSourceService()
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to initialize source service", err)
		return
	}
	c.Ctx.Request = c.Ctx.Request.WithContext(ctx)
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	projectID := c.Ctx.Input.Param(":projectid")
	sources, err := c.sourceService.GetAllSources(context.Background(), projectID)
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
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.sourceService.CreateSource(context.Background(), req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create source", err)
		return
	}

	// Track source creation event
	telemetry.TrackSourceCreation(context.Background(), source)

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	id := GetIDFromPath(&c.Controller)
	var req dto.UpdateSourceRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	projectID := c.Ctx.Input.Param(":projectid")
	var req models.UpdateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.sourceService.UpdateSource(context.Background(), id, req, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update source", err)
		return
	}

	// Update fields
	existingSource.Name = req.Name
	existingSource.Config = req.Config
	existingSource.Type = req.Type
	existingSource.Version = req.Version
	existingSource.UpdatedAt = time.Now()

	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		existingSource.UpdatedBy = user
	}

	// Find jobs linked to this source
	jobs, err := c.jobORM.GetBySourceID(existingSource.ID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch jobs for source %s", err))
		return
	}

	// Cancel workflows for those jobs
	for _, job := range jobs {
		err := cancelJobWorkflow(c.tempClient, job, projectID)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to cancel workflow for job %s", err))
			return
		}
	}

	// Persist update
	if err := c.sourceORM.Update(existingSource); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to update source %s", err))
		return
	}

	// Track sources status after update
	telemetry.TrackSourcesStatus(context.Background())
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id := GetIDFromPath(&c.Controller)
	resp, err := c.sourceService.DeleteSource(context.Background(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			respondWithError(&c.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete source", err)
		}
		return
	}
	utils.SuccessResponse(&c.Controller, resp)

	telemetry.TrackSourcesStatus(context.Background())
	utils.SuccessResponse(&c.Controller, &models.DeleteSourceResponse{
		Name: source.Name,
	})
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req dto.SourceTestConnectionRequest
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

// @router /sources/streams[post]
func (c *SourceHandler) GetSourceCatalog() {
	var req dto.StreamsRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	catalog, err := c.sourceService.GetSourceCatalog(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source catalog", err)
		return
	}
	utils.SuccessResponse(&c.Controller, catalog)
	// Use Temporal client to get the catalog
	var newStreams map[string]interface{}
	if c.tempClient != nil {
		newStreams, err = c.tempClient.GetCatalog(
			c.Ctx.Request.Context(),
			req.Type,
			req.Version,
			encryptedConfig,
			oldStreams,
			req.JobName,
		)
	}
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to get catalog: %v", err))
		return
	}
	utils.SuccessResponse(&c.Controller, newStreams)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	id := GetIDFromPath(&c.Controller)
	jobs, err := c.sourceService.GetSourceJobs(context.Background(), id)
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
	if sourceType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "source type is required")
		return
	}

	versions, _, err := utils.GetDriverImageTags(c.Ctx.Request.Context(), fmt.Sprintf("olakego/source-%s", sourceType), true)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to get Docker versions"
		if err.Error() == "source type is required" {
			status = http.StatusBadRequest
			msg = "Source type is required"
		}
		respondWithError(&c.Controller, status, msg, err)
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"version": versions})
}

// @router /project/:projectid/sources/spec [get]
func (c *SourceHandler) GetProjectSourceSpec() {
	_ = c.Ctx.Input.Param(":projectid")

	var req dto.SpecRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	var specOutput models.SpecOutput
	var err error

	specOutput, err = c.tempClient.FetchSpec(
		c.Ctx.Request.Context(),
		"",
		req.Type,
		req.Version,
	)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to get spec: %v", err))
		return
	}

	utils.SuccessResponse(&c.Controller, dto.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOutput.Spec,
	})
}
