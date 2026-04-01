package optimization

import (
	"net/http"

	"github.com/datazip-inc/olake-ui/server/utils"
)

const badRequestStatusCode = http.StatusBadRequest

func (h *Handler) GetCatalog() {
	catalogName, ok := h.requiredCatalog()
	if !ok {
		return
	}

	olakeConfig, err := h.opt.GetCatalog(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, upstreamStatus(err), err.Error(), err)
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

	// Convert config to JSON string
	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid config format", err)
		return
	}

	result, err := h.opt.CreateCatalog(h.Ctx.Request.Context(), configJSON)
	if err != nil {
		utils.ErrorResponse(&h.Controller, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(&h.Controller, result, nil)
}

// updates an existing catalog
func (h *Handler) UpdateCatalog() {
	_, ok := h.requiredCatalog()
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

	// Convert config to JSON string
	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid config format", err)
		return
	}

	result, err := h.opt.UpdateCatalog(h.Ctx.Request.Context(), configJSON)
	if err != nil {
		utils.ErrorResponse(&h.Controller, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(&h.Controller, result, nil)
}

// deletes a catalog
func (h *Handler) DeleteCatalog() {
	catalogName, ok := h.requiredCatalog()
	if !ok {
		return
	}

	result, err := h.opt.DeleteCatalogInOpt(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(&h.Controller, result, nil)
}

func (h *Handler) requiredCatalog() (string, bool) {
	catalog := h.Ctx.Input.Param(":catalog")
	if catalog == "" {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog name is not present in query params", nil)
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
