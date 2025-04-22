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

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	utils.SuccessResponse(&c.Controller, sources)
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
	projectID := c.Ctx.Input.Param(":projectid")
	if projectID != "" && projectID != "olake" {
		// Convert to uint if needed
		pid, err := strconv.ParseUint(projectID, 10, 64)
		if err == nil {
			source.ProjectID = uint(pid)
		}
	}

	// Set created by if user is logged in
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		source.CreatedBy = user
		source.UpdatedBy = user
	}

	if err := c.sourceORM.Create(source); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create source: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, models.CreateSourceResponse{
		Success: true,
		Message: "Source created successfully",
		Data:    req,
	})
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

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

	utils.SuccessResponse(&c.Controller, models.UpdateSourceResponse{
		Success: true,
		Message: "Source updated successfully",
		Data:    req,
	})
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}
	name, err := c.sourceORM.GetName(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}
	if err := c.sourceORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}

	utils.SuccessResponse(&c.Controller, models.DeleteSourceResponse{
		Success: true,
		Message: "Source deleted successfully",
		Data: struct {
			Name string `json:"name"`
		}{Name: name},
	})
}

// @router /sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// For now, always return success
	utils.SuccessResponse(&c.Controller, models.SourceTestConnectionResponse{
		Success: true,
		Message: "Connection successful",
		Data:    req,
	})

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

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	_, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Get source type from query parameter
	sourceType := c.GetString("type")
	if sourceType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Source type is required")
		return
	}

	// In a real implementation, we would query for available versions
	// based on the source type and project ID
	// For now, we'll return a mock response

	// Mock available versions (this would be replaced with actual DB query)
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}

	response := map[string]interface{}{
		"version": versions,
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/sources/spec [get]
func (c *SourceHandler) GetProjectSourceSpec() {
	// Get project ID from path (not used in current implementation)
	_ = c.Ctx.Input.Param(":projectid")
	// Will be used for multi-tenant filtering in the future
	var req models.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	// In a real implementation, we would fetch the specification
	// based on the source type, version and project ID
	// For now, we'll return a mock response

	// Mock specification (this would be replaced with actual data)
	mockSpec := "{ \"host\": \"string\", \"port\": \"integer\", \"username\": \"string\", \"password\": \"string\", \"database\": \"string\", \"ssl\": \"boolean\", \"timeout\": \"integer\" }"

	utils.SuccessResponse(&c.Controller, models.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    mockSpec,
	})
}
