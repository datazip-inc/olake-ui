package etl

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httpx"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @Summary Get project settings
// @Tags Project Settings
// @Description Retrieve the settings for a specific project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Success 200 {object} dto.JSONResponse{data=dto.ProjectSettingsResponse} "Project Settings fetched successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to retrieve project settings"
// @Router /api/v1/project/{projectid}/settings [get]
func (h *Handler) GetProjectSettings(c *gin.Context) {
	projectID, err := httpx.GetProjectID(c)
	if err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get project settings initiated project_id[%s]", projectID)
	settings, err := h.etl.GetProjectSettings(projectID)
	if err != nil {
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve project settings by project ID: %s", err), err)
		return
	}
	httpx.SuccessResponse(c, "Project Settings fetched successfully", settings)
}

// @Summary Update project settings
// @Tags Project Settings
// @Description Create or update the settings for a project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.UpsertProjectSettingsRequest true "project settings data"
// @Success 200 {object} dto.JSONResponse "Project Settings updated successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to update project settings"
// @Router /api/v1/project/{projectid}/settings [put]
func (h *Handler) UpsertProjectSettings(c *gin.Context) {
	projectID, err := httpx.GetProjectID(c)
	if err != nil {
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpsertProjectSettingsRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if req.ProjectID != projectID {
		err := fmt.Errorf("path project_id '%s' does not match body project_id '%s'", projectID, req.ProjectID)
		httpx.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Upsert project settings initiated project_id[%s]", projectID)
	if err := h.etl.UpsertProjectSettings(req); err != nil {
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to update project settings: %s", err), err)
		return
	}
	httpx.SuccessResponse(c, "Project Settings updated successfully", nil)
}
