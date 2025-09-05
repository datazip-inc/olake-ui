package handlers

import (
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

type DestHandler struct {
	web.Controller
	destORM    *database.DestinationORM
	jobORM     *database.JobORM
	userORM    *database.UserORM
	tempClient *temporal.Client
}

func (c *DestHandler) Prepare() {
	c.destORM = database.NewDestinationORM()
	c.jobORM = database.NewJobORM()
	c.userORM = database.NewUserORM()
	var err error
	c.tempClient, err = temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
	}
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	projectIDStr := c.Ctx.Input.Param(":projectid")
	destinations, err := c.destORM.GetAllByProjectID(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve destinations")
		return
	}
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

		setUsernames(&item.CreatedBy, &item.UpdatedBy, dest.CreatedBy, dest.UpdatedBy)

		jobs, err := c.jobORM.GetByDestinationID(dest.ID)
		var success bool
		item.Jobs, success = buildJobDataItems(jobs, err, projectIDStr, "destination", c.tempClient, &c.Controller)
		if !success {
			return // Error occurred in buildJobDataItems
		}

		destItems = append(destItems, item)
	}

	utils.SuccessResponse(&c.Controller, destItems)
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

	// Track destination creation event
	telemetry.TrackDestinationCreation(c.Ctx.Request.Context(), destination)
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	// Get destination ID from path
	id := GetIDFromPath(&c.Controller)

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

	// Track destinations status after update
	telemetry.TrackDestinationsStatus(c.Ctx.Request.Context())

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	// Get destination ID from path
	id := GetIDFromPath(&c.Controller)
	// Get the name for the response
	dest, err := c.destORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}

	jobs, err := c.jobORM.GetByDestinationID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get source by id")
	}
	for _, job := range jobs {
		job.Active = false
		if err := c.jobORM.Update(job); err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to deactivate jobs using this destination")
			return
		}
	}
	if err := c.destORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete destination")
		return
	}

	// Track destinations status after deletion
	telemetry.TrackDestinationsStatus(c.Ctx.Request.Context())

	utils.SuccessResponse(&c.Controller, &models.DeleteDestinationResponse{
		Name: dest.Name,
	})
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	var req models.DestinationTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	if req.Type == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "valid destination type is required")
		return
	}

	if req.Version == "" || req.Version == "latest" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "valid destination version required")
		return
	}

	// Determine driver and available tags
	version := req.Version
	driver := req.Source
	if driver == "" {
		var err error
		_, driver, err = utils.GetDriverImageTags(c.Ctx.Request.Context(), "", true)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get valid driver image tags: %s", err))
			return
		}
	}

	encryptedConfig, err := utils.Encrypt(req.Config)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to encrypt destination config: "+err.Error())
		return
	}

	result, err := c.tempClient.TestConnection(c.Ctx.Request.Context(), "destination", driver, version, encryptedConfig)
	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /destinations/:id/jobs [get]
func (c *DestHandler) GetDestinationJobs() {
	id := GetIDFromPath(&c.Controller)
	// Check if destination exists
	_, err := c.destORM.GetByID(id)
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
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"jobs": jobs,
	})
}

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	// Get destination type from query parameter
	destType := c.GetString("type")
	if destType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	// get available driver versions
	versions, _, err := utils.GetDriverImageTags(c.Ctx.Request.Context(), "", true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to fetch driver versions: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version": versions,
	})
}

// @router /project/:projectid/destinations/spec [post]
func (c *DestHandler) GetDestinationSpec() {
	_ = c.Ctx.Input.Param(":projectid")

	var req models.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	var spec models.SpecOutput
	var err error
	// TODO: make destinationType consistent. Only use parquet and iceberg.
	destinationType := "iceberg"
	if req.Type == "s3" {
		destinationType = "parquet"
	}
	if c.tempClient != nil {
		spec, err = c.tempClient.FetchSpec(
			c.Ctx.Request.Context(),
			destinationType,
			"postgres",
			req.Version,
		)
	}
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to get spec: %v", err))
		return
	}

	utils.SuccessResponse(&c.Controller, models.SpecResponse{
		Version:    req.Version,
		Type:       req.Type,
		JSONSchema: spec.JSONSchema,
		UISchema:   spec.UISchema,
	})
}
