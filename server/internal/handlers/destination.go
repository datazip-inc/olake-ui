package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/utils"
)

type DestHandler struct {
	web.Controller
	destService *services.DestinationService
}

func (c *DestHandler) Prepare() {
	var err error
	c.destService, err = services.NewDestinationService()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to initialize destination service")
		return
	}
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	projectID := c.Ctx.Input.Param(":projectid")

	destinations, err := c.destService.GetAllDestinations(context.Background(), projectID)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get destinations", err)
		return
	}

	utils.SuccessResponse(&c.Controller, destinations)
}

// @router /project/:projectid/destinations [post]
func (c *DestHandler) CreateDestination() {
	projectID := c.Ctx.Input.Param(":projectid")
	if projectID == "" {
		respondWithError(&c.Controller, http.StatusBadRequest, "project ID is required", nil)
		return
	}

	var req dto.CreateDestinationRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	// Convert request to Destination model
	destination := &models.Destination{
		Name:      req.Name,
		DestType:  req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectIDStr,
	}

	if err := c.destService.CreateDestination(context.Background(), req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create destination", err)
		return
	}

	// Track destination creation event
	telemetry.TrackDestinationCreation(context.Background(), destination)
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	id := GetIDFromPath(&c.Controller)

	var req dto.UpdateDestinationRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	projectID := c.Ctx.Input.Param(":projectid")
	var req models.UpdateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)

	if err := c.destService.UpdateDestination(context.Background(), id, req, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update destination", err)
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

	// Find jobs linked to this source
	jobs, err := c.jobORM.GetByDestinationID(existingDest.ID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch jobs for destination %s", err))
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

	// persist update
	if err := c.destORM.Update(existingDest); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to update destination %s", err))
		return
	}

	// Track destinations status after update
	telemetry.TrackDestinationsStatus(context.Background())

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	id := GetIDFromPath(&c.Controller)

	response, err := c.destService.DeleteDestination(context.Background(), id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete destination", err)
		return
	}

	utils.SuccessResponse(&c.Controller, response)
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
	telemetry.TrackDestinationsStatus(context.Background())

	utils.SuccessResponse(&c.Controller, &models.DeleteDestinationResponse{
		Name: dest.Name,
	})
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	var req dto.DestinationTestConnectionRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	var req models.DestinationTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	result, err := c.destService.TestConnection(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Failed to test connection", err)
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

	jobs, err := c.destService.GetDestinationJobs(context.Background(), id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get destination jobs", err)
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"jobs": jobs,
	})
}

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	destType := c.GetString("type")

	versions, err := c.destService.GetDestinationVersions(context.Background(), destType)
	// get available driver versions
	versions, _, err := utils.GetDriverImageTags(c.Ctx.Request.Context(), "", true)
	if err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Failed to get destination versions", err)
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version": versions,
	})
}

// @router /project/:projectid/destinations/spec [post]
func (c *DestHandler) GetDestinationSpec() {
	_ = c.Ctx.Input.Param(":projectid")

	var req dto.SpecRequest
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	var specOutput models.SpecOutput
	var err error
	// TODO: make destinationType consistent. Only use parquet and iceberg.
	destinationType := "iceberg"
	if req.Type == "s3" {
		destinationType = "parquet"
	}
	version := req.Version

	if req.Type == "" {
		respondWithError(&c.Controller, http.StatusBadRequest, "Destination type is required", nil)
		return
	}

	if req.Version == "" {
		respondWithError(&c.Controller, http.StatusBadRequest, "Destination version is required", nil)
	// Determine driver and available tags
	_, driver, err := utils.GetDriverImageTags(c.Ctx.Request.Context(), "", true)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get valid driver image tags: %s", err))
		return
	}

	if c.tempClient != nil {
		specOutput, err = c.tempClient.FetchSpec(
			c.Ctx.Request.Context(),
			destinationType,
			driver,
			version,
		)
	}
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to get spec: %v", err))
		return
	}

	utils.SuccessResponse(&c.Controller, models.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    specOutput.Spec,
	})
}
