package optimization

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httpx"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

const badRequestStatusCode = http.StatusBadRequest

func (h *Handler) GetCatalog(c *gin.Context) {
	catalogName, ok := h.requiredCatalog(c)
	if !ok {
		return
	}

	olakeConfig, err := h.opt.GetCatalog(c.Request.Context(), catalogName)
	if err != nil {
		httpx.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	httpx.SuccessResponse(c, "catalog details retrieved successfully", olakeConfig)
}

func (h *Handler) CreateCatalog(c *gin.Context) {
	var req map[string]interface{}
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if req == nil {
		httpx.ErrorResponse(c, badRequestStatusCode, "catalog config is required", nil)
		return
	}

	// Convert config to JSON string
	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		httpx.ErrorResponse(c, badRequestStatusCode, "invalid config format", err)
		return
	}

	result, err := h.opt.CreateCatalog(c.Request.Context(), configJSON)
	if err != nil {
		httpx.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	httpx.SuccessResponse(c, result, nil)
}

// updates an existing catalog
func (h *Handler) UpdateCatalog(c *gin.Context) {
	_, ok := h.requiredCatalog(c)
	if !ok {
		return
	}

	var req map[string]interface{}
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if req == nil {
		httpx.ErrorResponse(c, badRequestStatusCode, "catalog config is required", nil)
		return
	}

	// Convert config to JSON string
	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		httpx.ErrorResponse(c, badRequestStatusCode, "invalid config format", err)
		return
	}

	result, err := h.opt.UpdateCatalog(c.Request.Context(), configJSON)
	if err != nil {
		httpx.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	httpx.SuccessResponse(c, fmt.Sprintf("catalog %s updated successfully", result), nil)
}

// deletes a catalog
func (h *Handler) DeleteCatalog(c *gin.Context) {
	catalogName, ok := h.requiredCatalog(c)
	if !ok {
		return
	}

	result, err := h.opt.DeleteCatalog(c.Request.Context(), catalogName)
	if err != nil {
		httpx.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	httpx.SuccessResponse(c, fmt.Sprintf("catalog %s deleted successfully", result), nil)
}

func (h *Handler) requiredCatalog(c *gin.Context) (string, bool) {
	catalog := c.Param("catalog")
	if catalog == "" {
		httpx.ErrorResponse(c, badRequestStatusCode, "catalog name is required", nil)
		return "", false
	}
	return catalog, true
}
