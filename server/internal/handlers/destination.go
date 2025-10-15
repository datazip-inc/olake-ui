package handlers

import (
	"context"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/utils"
)

type DestHandler struct {
	web.Controller
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	projectID := c.Ctx.Input.Param(":projectid")
	logger.Info("Get all destinations initiated - project_id=%s", projectID)

	items, err := DestSvc().GetAllDestinations(c.Ctx.Request.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get destinations", err)
		return
	}
	utils.SuccessResponse(&c.Controller, items)
}

// @router /project/:projectid/destinations [post]
func (c *DestHandler) CreateDestination() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.CreateDestinationRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Create destination initiated - project_id=%s destination_type=%s destination_name=%s user_id=%v",
		projectID, req.Type, req.Name, userID)

	if err := DestSvc().CreateDestination(context.Background(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to create destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	id := GetIDFromPath(&c.Controller)
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.UpdateDestinationRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)
	logger.Info("Update destination initiated - project_id=%s destination_id=%d destination_type=%s user_id=%v",
		projectID, id, req.Type, userID)

	if err := DestSvc().UpdateDestination(context.Background(), id, projectID, &req, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	id := GetIDFromPath(&c.Controller)
	logger.Info("Delete destination initiated - destination_id=%d", id)

	resp, err := DestSvc().DeleteDestination(context.Background(), id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete destination", err)
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	var req dto.DestinationTestConnectionRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Test destination connection initiated - destination_type=%s destination_version=%s", req.Type, req.Version)

	result, logs, err := DestSvc().TestConnection(context.Background(), &req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Failed to test connection", err)
		return
	}

	utils.SuccessResponse(&c.Controller, dto.TestConnectionResponse{
		ConnectionResult: result,
		Logs:             logs,
	})
}

// @router /destinations/:id/jobs [get]
func (c *DestHandler) GetDestinationJobs() {
	id := GetIDFromPath(&c.Controller)
	logger.Info("Get destination jobs initiated - destination_id=%d", id)

	jobs, err := DestSvc().GetDestinationJobs(context.Background(), id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get destination jobs", err)
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"jobs": jobs})
}

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	projectID := c.Ctx.Input.Param(":projectid")
	destType := c.GetString("type")
	logger.Info("Get destination versions initiated - project_id=%s destination_type=%s", projectID, destType)

	versions, err := DestSvc().GetDestinationVersions(context.Background(), destType)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Failed to get destination versions", err)
		return
	}
	utils.SuccessResponse(&c.Controller, versions)
}

// @router /project/:projectid/destinations/spec [post]
func (c *DestHandler) GetDestinationSpec() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req dto.SpecRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Get destination spec initiated - project_id=%s destination_type=%s destination_version=%s",
		projectID, req.Type, req.Version)

	resp, err := DestSvc().GetDestinationSpec(c.Ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get destination spec", err)
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}
