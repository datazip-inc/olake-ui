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

type DestHandler struct {
	web.Controller
	destORM *database.DestinationORM
}

func (c *DestHandler) Prepare() {
	c.destORM = database.NewDestinationORM()
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	// Get project ID from path
	//use olake project id when is needed
	//projectIDStr := c.Ctx.Input.Param(":projectid")

	destinations, err := c.destORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve destinations")
		return
	}

	// Filter destinations by project ID in memory
	// In a real implementation, this would be done in the database query
	// var filteredDestinations []*models.Destination
	// for _, dest := range destinations {
	// 	if dest.ProjectID == projectIDStr {
	// 		filteredDestinations = append(filteredDestinations, dest)
	// 	}
	// }

	// Format response data
	destItems := make([]models.DestinationDataItem, 0, len(destinations))
	for _, dest := range destinations {
		item := models.DestinationDataItem{
			ID:        dest.ID,
			Name:      dest.Name,
			Type:      dest.DestType,
			Version:   dest.Version,
			Config:    dest.Config,
			CreatedAt: dest.CreatedAt.Format(time.RFC3339),
			UpdatedAt: dest.UpdatedAt.Format(time.RFC3339),
		}

		// Add creator username if available
		if dest.CreatedBy != nil {
			item.CreatedBy = dest.CreatedBy.Username
		}

		// Add updater username if available
		if dest.UpdatedBy != nil {
			item.UpdatedBy = dest.UpdatedBy.Username
		}

		destItems = append(destItems, item)
	}

	response := models.JSONResponse{
		Success: true,
		Message: "Destinations retrieved successfully",
		Data:    destItems,
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/destinations [post]
func (c *DestHandler) CreateDestination() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")

	var req models.CreateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert request to Destination model
	destination := &models.Destination{
		Name:      req.Name,
		DestType:  req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectIDStr,
	}

	// Set created by if user is logged in
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		destination.CreatedBy = user
		destination.UpdatedBy = user
	}

	if err := c.destORM.Create(destination); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create destination: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, models.CreateDestinationResponse{
		Success: true,
		Message: "Destination created successfully",
		Data:    req,
	})
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		fmt.Println("Error converting project ID to int:", err)
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Project ID will be used for permission checking in future implementations
	_ = projectID

	// Get destination ID from path
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	var req models.UpdateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get existing destination
	existingDest, err := c.destORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}

	// Update fields
	existingDest.Name = req.Name
	existingDest.DestType = req.Type
	existingDest.Version = req.Version
	existingDest.Config = req.Config
	existingDest.UpdatedAt = time.Now()

	// Update user who made changes
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		existingDest.UpdatedBy = user
	}

	if err := c.destORM.Update(existingDest); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update destination")
		return
	}

	utils.SuccessResponse(&c.Controller, models.UpdateDestinationResponse{
		Success: true,
		Message: "Destination updated successfully",
		Data:    req,
	})
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		fmt.Println("Error converting project ID to int:", err)
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Project ID will be used for permission checking in future implementations
	_ = projectID

	// Get destination ID from path
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	// Get the name for the response
	name, err := c.destORM.GetName(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}

	if err := c.destORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete destination")
		return
	}

	response := models.DeleteDestinationResponse{
		Success: true,
		Message: "Destination deleted successfully",
		Data: struct {
			Name string `json:"name"`
		}{
			Name: name,
		},
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	// Get project ID from path (not used in current implementation)
	_ = c.Ctx.Input.Param(":projectid")
	// Will be used for multi-tenant filtering in the future
	var req models.DestinationTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Type == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	if req.Version == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination version is required")
		return
	}

	// For now, always return success
	utils.SuccessResponse(&c.Controller, models.DestinationTestConnectionResponse{
		Success: true,
		Message: "Connection successful",
		Data:    req,
	})
}

// @router /destinations/:id/jobs [get]
func (c *DestHandler) GetDestinationJobs() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	// Check if destination exists
	_, err = c.destORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}

	// Create a job ORM and get jobs by destination ID
	jobORM := database.NewJobORM()
	jobs, err := jobORM.GetByDestinationID(id)
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

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	_, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Get destination type from query parameter
	destType := c.GetString("type")
	if destType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	// In a real implementation, we would query for available versions
	// based on the destination type and project ID
	// For now, we'll return a mock response

	// Mock available versions (this would be replaced with actual DB query)
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}

	response := map[string]interface{}{
		"version": versions,
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/destinations/spec [get]
func (c *DestHandler) GetDestinationSpec() {
	// Get project ID from path (not used in current implementation)
	_ = c.Ctx.Input.Param(":projectid")
	// Will be used for multi-tenant filtering in the future

	var req models.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Type == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	if req.Version == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination version is required")
		return
	}

	// In a real implementation, we would fetch the specification
	// based on the destination type, version and project ID
	// For now, we'll return a mock response

	// Mock specification (this would be replaced with actual data)
	var mockSpec = `{ "auth_type": "string", "aws_access_key_id": "string", "aws_secret_key": "string", "bucket_name": "string", "bucket_path": "string", "region": "string" }`

	utils.SuccessResponse(&c.Controller, models.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    mockSpec,
	})
}
