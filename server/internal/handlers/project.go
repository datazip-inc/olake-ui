package handlers

import (
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @router /project/:projectid/settings [get]
func (h *Handler) GetProjectSettings() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get project settings initiated project_id[%s]", projectID)

	settings, err := h.etl.GetProjectSettings(projectID)
	if err != nil {
		status := http.StatusInternalServerError
		utils.ErrorResponse(&h.Controller, status, fmt.Sprintf("failed to retrieve project settings by project ID: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, "Project Settings fetched successfully", settings)
}

// @router /project/:projectid/settings [put]
func (h *Handler) UpdateProjectSettings() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpdateProjectSettingsRequest

	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Update project settings initiated project_id[%s]", projectID)

	if err := h.etl.UpdateProjectSettings(projectID, req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update project settings: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Project Settings updated successfully", nil)
}
