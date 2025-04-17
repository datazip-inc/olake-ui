package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/utils"
)

type SourceHandler struct {
	web.Controller
	sourceORM *database.SourceORM
}

func (c *SourceHandler) Prepare() {
	c.sourceORM = database.NewSourceORM()
}

// @router /sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	utils.SuccessResponse(&c.Controller, sources)
}

// @router /sources [post]
func (c *SourceHandler) CreateSource() {
	var req models.Source
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		req.CreatedBy = user
		req.UpdatedBy = user
	}

	if err := c.sourceORM.Create(&req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create source: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	var req models.Source
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
	existingSource.ProjectID = req.ProjectID
	existingSource.Config = req.Config
	existingSource.SourceType = req.SourceType
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

	utils.SuccessResponse(&c.Controller, existingSource)
}

// @router /sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	if err := c.sourceORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}

	c.Ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
}

// @router /sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// For now, always return success
	response := map[string]string{
		"status": "success",
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// @router /sources/:source_type/spec [get]
func (c *SourceHandler) GetSourceTypeSpec() {
	sourceType := c.Ctx.Input.Param(":source_type")
	if sourceType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Source type is required")
		return
	}

	// Return empty spec object for now
	response := map[string]interface{}{
		"spec": map[string]interface{}{},
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// @router /sources/:id/catalog [get]
func (c *SourceHandler) GetSourceCatalog() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	// Get existing source
	_, err = c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Return empty catalog object for now
	// TODO: Implement actual catalog generation logic
	response := map[string]interface{}{
		"catalog": map[string]interface{}{},
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	// Check if source exists
	_, err = c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Create a job ORM and get jobs by source ID
	jobORM := database.NewJobORM()
	jobs, err := jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs")
		return
	}

	// Format as required by API contract
	response := map[string]interface{}{
		"jobs": jobs,
	}

	c.Data["json"] = response
	c.ServeJSON()
}
