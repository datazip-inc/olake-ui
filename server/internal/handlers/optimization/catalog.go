package optimization

import (
	"net/http"

	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

const badRequestStatusCode = http.StatusBadRequest

func (h *Handler) GetCatalog() {
	catalogName, ok := h.requiredCatalog()
	if !ok {
		return
	}

	logger.Debugf("Get catalog details initiated catalog[%s]", catalogName)

	olakeConfig, err := h.opt.GetCatalogAsOLakeConfig(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to get catalog details", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "catalog details retrieved successfully", olakeConfig)
}

func (h *Handler) CreateCatalog() {
	var req map[string]interface{}
	if !h.bindJSON(&req) {
		return
	}

	if req == nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog config is required", nil)
		return
	}

	logger.Debugf("Create catalog initiated")

	// Convert config to JSON string
	configJSON, err := utils.ToJSON(req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid config format", err)
		return
	}

	result, err := h.opt.CreateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON, false)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to create catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, nil)
}

// updates an existing catalog
func (h *Handler) UpdateCatalog() {
	catalogName, ok := h.requiredCatalog()
	if !ok {
		return
	}

	var req map[string]interface{}
	if !h.bindJSON(&req) {
		return
	}

	if req == nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog config is required", nil)
		return
	}

	logger.Debugf("Update catalog initiated catalog[%s]", catalogName)

	// Convert config to JSON string
	configJSON, err := utils.ToJSON(req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid config format", err)
		return
	}

	result, err := h.opt.UpdateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to update catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, nil)
}

// deletes a catalog
func (h *Handler) DeleteCatalog() {
	catalogName, ok := h.requiredCatalog()
	if !ok {
		return
	}

	logger.Debugf("Delete catalog initiated catalog[%s]", catalogName)

	result, err := h.opt.DeleteCatalog(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to delete catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, nil)
}

func (h *Handler) requiredCatalog() (string, bool) {
	catalog := h.Ctx.Input.Param(":catalog")
	if catalog == "" {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog name is required", nil)
		return "", false
	}
	return catalog, true
}

func (h *Handler) bindJSON(dst interface{}) bool {
	if err := h.Ctx.BindJSON(dst); err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid request body", err)
		return false
	}
	return true
}