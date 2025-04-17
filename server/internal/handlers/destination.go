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

// @router /destinations [get]
func (c *DestHandler) GetAllDestinations() {
	destinations, err := c.destORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve destinations")
		return
	}

	utils.SuccessResponse(&c.Controller, destinations)
}

// @router /destinations [post]
func (c *DestHandler) CreateDestination() {
	var req models.Destination
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

	if err := c.destORM.Create(&req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create destination: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	var req models.Destination
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
	existingDest.ProjectID = req.ProjectID
	existingDest.Config = req.Config
	existingDest.DestType = req.DestType
	existingDest.UpdatedAt = time.Now()

	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		existingDest.UpdatedBy = user
	}

	if err := c.destORM.Update(existingDest); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update destination")
		return
	}

	utils.SuccessResponse(&c.Controller, existingDest)
}

// @router /destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	if err := c.destORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete destination")
		return
	}

	c.Ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
}

// @router /destinations/test [post]
func (c *DestHandler) TestConnection() {
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

// @router /destinations/:dest_type/spec [get]
func (c *DestHandler) GetDestTypeSpec() {
	destType := c.Ctx.Input.Param(":dest_type")
	if destType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	// Return empty spec object for now
	response := map[string]interface{}{
		"spec": map[string]interface{}{},
	}

	c.Data["json"] = response
	c.ServeJSON()
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
