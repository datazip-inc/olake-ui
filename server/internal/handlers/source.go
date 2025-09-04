package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/internal/temporal"
	"github.com/datazip/olake-frontend/server/utils"
)

type SourceHandler struct {
	web.Controller
	sourceORM  *database.SourceORM
	userORM    *database.UserORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func (c *SourceHandler) Prepare() {
	c.sourceORM = database.NewSourceORM()
	c.userORM = database.NewUserORM()
	c.jobORM = database.NewJobORM()

	// Initialize Temporal client
	var err error
	c.tempClient, err = temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
	}
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	projectIDStr := c.Ctx.Input.Param(":projectid")
	sourceItems := make([]models.SourceDataItem, 0, len(sources))

	for _, source := range sources {
		item := models.SourceDataItem{
			ID:        source.ID,
			Name:      source.Name,
			Type:      source.Type,
			Version:   source.Version,
			Config:    source.Config,
			CreatedAt: source.CreatedAt.Format(time.RFC3339),
			UpdatedAt: source.UpdatedAt.Format(time.RFC3339),
		}

		setUsernames(&item.CreatedBy, &item.UpdatedBy, source.CreatedBy, source.UpdatedBy)

		jobs, err := c.jobORM.GetBySourceID(source.ID)
		var success bool
		item.Jobs, success = buildJobDataItems(jobs, err, projectIDStr, "source", c.tempClient, &c.Controller)
		if !success {
			return // Error occurred in buildJobDataItems
		}

		sourceItems = append(sourceItems, item)
	}

	utils.SuccessResponse(&c.Controller, sourceItems)
}

// @router /project/:projectid/sources [post]
func (c *SourceHandler) CreateSource() {
	var req models.CreateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert request to Source model
	source := &models.Source{
		Name:    req.Name,
		Type:    req.Type,
		Version: req.Version,
		Config:  req.Config,
	}

	// Get project ID if needed
	source.ProjectID = c.Ctx.Input.Param(":projectid")

	// Set created by if user is logged in
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user, err := c.userORM.GetByID(userID.(int))
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get user")
			return
		}
		source.CreatedBy = user
		source.UpdatedBy = user
	}
	if err := c.sourceORM.Create(source); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create source: %s", err))
		return
	}

	// Track source creation event
	telemetry.TrackSourceCreation(c.Ctx.Request.Context(), source)

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	id := GetIDFromPath(&c.Controller)
	var req models.UpdateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	// Get existing source
	existingSource, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
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

	if err := c.sourceORM.Update(existingSource); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update source")
		return
	}

	// Track sources status after update
	telemetry.TrackSourcesStatus(c.Ctx.Request.Context())
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id := GetIDFromPath(&c.Controller)
	source, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Get all jobs using this source
	jobs, err := c.jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get jobs for source")
		return
	}

	// Deactivate all jobs using this source
	for _, job := range jobs {
		job.Active = false
		if err := c.jobORM.Update(job); err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to deactivate jobs using this source")
			return
		}
	}

	// Delete the source
	if err := c.sourceORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}

	telemetry.TrackSourcesStatus(c.Ctx.Request.Context())
	utils.SuccessResponse(&c.Controller, &models.DeleteSourceResponse{
		Name: source.Name,
	})
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to encrypt config")
		return
	}
	result, err := c.tempClient.TestConnection(context.Background(), "config", req.Type, req.Version, encryptedConfig)
	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /sources/streams[post]
func (c *SourceHandler) GetSourceCatalog() {
	var req models.StreamsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	oldStreams := ""
	// Load job details if JobID is provided
	if req.JobID >= 0 {
		job, err := c.jobORM.GetByID(req.JobID, true)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
			return
		}
		oldStreams = job.StreamsConfig
	}
	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to encrypt config")
		return
	}
	// Use Temporal client to get the catalog
	var newStreams map[string]interface{}
	if c.tempClient != nil {
		newStreams, err = c.tempClient.GetCatalog(
			c.Ctx.Request.Context(),
			req.Type,
			req.Version,
			encryptedConfig,
			oldStreams,
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
	// Check if source exists
	_, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Create a job ORM and get jobs by source ID
	jobs, err := c.jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get jobs by source ID")
		return
	}
	// Format as required by API contract
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"jobs": jobs,
	})
}

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	sourceType := c.GetString("type")
	if sourceType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "source type is required")
		return
	}

	versions, _, err := utils.GetDriverImageTags(c.Ctx.Request.Context(), fmt.Sprintf("olakego/source-%s", sourceType), true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to fetch driver versions: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version": versions,
	})
}

// @router /project/:projectid/sources/spec [get]
func (c *SourceHandler) GetProjectSourceSpec() {
	_ = c.Ctx.Input.Param(":projectid")

	var req models.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
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

	utils.SuccessResponse(&c.Controller, models.SpecResponse{
		Version:    req.Version,
		Type:       req.Type,
		JSONSchema: specOutput.JSONSchema,
		UISchema:   specOutput.UISchema,
	})
}
