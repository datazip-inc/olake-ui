package etlhandlers

import (
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
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
func (h *Handler) GetProjectSettings() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get project settings initiated project_id[%s]", projectID)

	settings, err := h.etl.GetProjectSettings(projectID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve project settings by project ID: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, "Project Settings fetched successfully", settings)
}

// @Summary Update project settings
// @Tags Project Settings
// @Description Create or update the settings for a project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.UpsertProjectSettingsRequest true "project settings data"
// @Success 200 {object} dto.JSONResponse "Project Settings updated successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to update project settings"
// @Router /api/v1/project/{projectid}/settings [put]
func (h *Handler) UpsertProjectSettings() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpsertProjectSettingsRequest

	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Update project settings initiated project_id[%s]", projectID)

	if err := h.etl.UpsertProjectSettings(req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update project settings: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Project Settings updated successfully", nil)
}
