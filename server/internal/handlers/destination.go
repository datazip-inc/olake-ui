package handlers

import (
	"context"
	"encoding/json"
	"net/http"

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
	svc, err := services.NewDestinationService()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to initialize destination service")
		return
	}
	c.destService = svc
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	projectID := c.Ctx.Input.Param(":projectid")
	items, err := c.destService.GetAllDestinations(c.Ctx.Request.Context(), projectID)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get destinations", err)
		return
	}
	utils.SuccessResponse(&c.Controller, items)
}

// @router /project/:projectid/destinations [post]
func (c *DestHandler) CreateDestination() {
	projectID := c.Ctx.Input.Param(":projectid")
	if projectID == "" {
		respondWithError(&c.Controller, http.StatusBadRequest, "project ID is required", nil)
		return
	}

	var req dto.CreateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	if err := c.destService.CreateDestination(context.Background(), req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	id := GetIDFromPath(&c.Controller)
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.UpdateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)

	if err := c.destService.UpdateDestination(context.Background(), id, projectID, req, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	id := GetIDFromPath(&c.Controller)

	resp, err := c.destService.DeleteDestination(context.Background(), id)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	var req dto.DestinationTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	result, err := c.destService.TestConnection(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Failed to test connection", err)
		return
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
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"jobs": jobs})
}

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	destType := c.GetString("type")

	versions, err := c.destService.GetDestinationVersions(context.Background(), destType)
	if err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Failed to get destination versions", err)
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"version": versions})
}

// @router /project/:projectid/destinations/spec [post]
func (c *DestHandler) GetDestinationSpec() {
	var req dto.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	resp, err := c.destService.GetDestinationSpec(c.Ctx.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}
