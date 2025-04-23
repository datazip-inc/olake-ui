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
	userORM   *database.UserORM
}

func (c *SourceHandler) Prepare() {
	c.sourceORM = database.NewSourceORM()
	c.userORM = database.NewUserORM()
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	// Format response data
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

		// Add creator username if available
		if source.CreatedBy != nil {
			item.CreatedBy = source.CreatedBy.Username
		}

		// Add updater username if available
		if source.UpdatedBy != nil {
			item.UpdatedBy = source.UpdatedBy.Username
		}

		sourceItems = append(sourceItems, item)
	}

	response := models.JSONResponse{
		Success: true,
		Message: "Sources retrieved successfully",
		Data:    sourceItems,
	}

	utils.SuccessResponse(&c.Controller, response)
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
	projectIDStr := c.Ctx.Input.Param(":projectid")
	source.ProjectID = projectIDStr

	// Set created by if user is logged in
	userID := c.GetSession(constants.SessionUserID)
	fmt.Printf("userID: %+v\n", userID)
	if userID != nil {
		user, err := c.userORM.GetByID(userID.(int))
		fmt.Printf("user: %+v\n", user)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get user")
			return
		}
		source.CreatedBy = user
		source.UpdatedBy = user
	}
	fmt.Printf("source: %+v\n", source)

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

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// For now, always return success
	utils.SuccessResponse(&c.Controller, models.JSONResponse{
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
	response := map[string]interface{}{"catalog": map[string]interface{}{"selected_streams": map[string]interface{}{"incr": []map[string]interface{}{{"split_column": "", "partition_regex": "", "stream_name": "incr3"}, {"split_column": "", "partition_regex": "", "stream_name": "incr4"}, {"split_column": "", "partition_regex": "", "stream_name": "incr1"}, {"split_column": "", "partition_regex": "", "stream_name": "incr2"}, {"split_column": "", "partition_regex": "", "stream_name": "incr"}}}, "streams": []map[string]interface{}{{"stream": map[string]interface{}{"name": "incr3", "namespace": "incr", "type_schema": map[string]interface{}{"properties": map[string]interface{}{"_id": map[string]interface{}{"type": []string{"string"}}, "address": map[string]interface{}{"type": []string{"string"}}, "age": map[string]interface{}{"type": []string{"integer"}}, "height": map[string]interface{}{"type": []string{"number"}}, "name": map[string]interface{}{"type": []string{"string"}}}}, "supported_sync_modes": []string{"full_refresh", "cdc"}, "source_defined_primary_key": []string{"_id"}, "available_cursor_fields": []string{}, "sync_mode": "cdc"}}, {"stream": map[string]interface{}{"name": "incr4", "namespace": "incr", "type_schema": map[string]interface{}{}, "supported_sync_modes": []string{"full_refresh", "cdc"}, "source_defined_primary_key": []string{"_id"}, "available_cursor_fields": []string{}, "sync_mode": "cdc"}}, {"stream": map[string]interface{}{"name": "incr1", "namespace": "incr", "type_schema": map[string]interface{}{"properties": map[string]interface{}{"_id": map[string]interface{}{"type": []string{"string"}}, "address": map[string]interface{}{"type": []string{"string"}}, "age": map[string]interface{}{"type": []string{"string"}}, "favo": map[string]interface{}{"type": []string{"string"}}, "height": map[string]interface{}{"type": []string{"string", "integer", "boolean", "number"}}, "last_modified": map[string]interface{}{"type": []string{"object"}}, "name": map[string]interface{}{"type": []string{"string"}}, "town": map[string]interface{}{"type": []string{"string"}}}}, "supported_sync_modes": []string{"cdc", "full_refresh"}, "source_defined_primary_key": []string{"_id"}, "available_cursor_fields": []string{}, "sync_mode": "cdc"}}, {"stream": map[string]interface{}{"name": "incr2", "namespace": "incr", "type_schema": map[string]interface{}{"properties": map[string]interface{}{"_id": map[string]interface{}{"type": []string{"string"}}, "address": map[string]interface{}{"type": []string{"string"}}, "age": map[string]interface{}{"type": []string{"integer"}}, "height": map[string]interface{}{"type": []string{"number"}}, "name": map[string]interface{}{"type": []string{"string"}}}}, "supported_sync_modes": []string{"full_refresh", "cdc"}, "source_defined_primary_key": []string{"_id"}, "available_cursor_fields": []string{}, "sync_mode": "cdc"}}, {"stream": map[string]interface{}{"name": "incr", "namespace": "incr", "type_schema": map[string]interface{}{"properties": map[string]interface{}{"_id": map[string]interface{}{"type": []string{"string"}}, "time": map[string]interface{}{"type": []string{"integer"}}, "time_here": map[string]interface{}{"type": []string{"integer"}}}}, "supported_sync_modes": []string{"full_refresh", "cdc"}, "source_defined_primary_key": []string{"_id"}, "available_cursor_fields": []string{}, "sync_mode": "cdc"}}}}}

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
