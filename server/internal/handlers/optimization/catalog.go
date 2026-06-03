package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
)


const badRequestStatusCode = http.StatusBadRequest

func (h *Handler) GetCatalog(c *gin.Context) {
	catalogName, ok := h.requiredCatalog(c)
	if !ok {
		return
	}

	olakeConfig, err := h.opt.GetCatalog(c.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, "catalog details retrieved successfully", olakeConfig)
}

func (h *Handler) TestCatalogConnection(c *gin.Context) {
	var req map[string]interface{}
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), "invalid request body for catalog test connection", err)
		return
	}

	if req == nil {
		utils.ErrorResponse(c, badRequestStatusCode, "catalog config is required for test connection", nil)
		return
	}

	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		utils.ErrorResponse(c, badRequestStatusCode, "invalid config format for catalog test connection", err)
		return
	}

	// checks if there is any changes in the config
	result, err := h.opt.TestCatalogConnection(c.Request.Context(), configJSON)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, "catalog connection tested successfully", result)
}

func (h *Handler) CreateCatalog(c *gin.Context) {
	var req map[string]interface{}
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), "invalid request body for catalog creation", err)
		return
	}

	if req == nil {
		utils.ErrorResponse(c, badRequestStatusCode, "catalog config is required during creation", nil)
		return
	}

	// Convert config to JSON string
	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		utils.ErrorResponse(c, badRequestStatusCode, "invalid config format for create catalog", err)
		return
	}

	result, err := h.opt.CreateCatalog(c.Request.Context(), configJSON)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, result, nil)
}

// updates an existing catalog
func (h *Handler) UpdateCatalog(c *gin.Context) {
	_, ok := h.requiredCatalog(c)
	if !ok {
		return
	}

	var req map[string]interface{}
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), "invalid request body for updating catalog", err)
		return
	}

	if req == nil {
		utils.ErrorResponse(c, badRequestStatusCode, "catalog config is required during updation", nil)
		return
	}

	// Convert config to JSON string
	configJSON, err := utils.MarshalToString(req)
	if err != nil {
		utils.ErrorResponse(c, badRequestStatusCode, "invalid config format for updating catalog", err)
		return
	}

	result, err := h.opt.UpdateCatalog(c.Request.Context(), configJSON)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, fmt.Sprintf("catalog %s updated successfully", result), nil)
}

// deletes a catalog
func (h *Handler) DeleteCatalog(c *gin.Context) {
	catalogName, ok := h.requiredCatalog(c)
	if !ok {
		return
	}

	result, err := h.opt.DeleteCatalog(c.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, fmt.Sprintf("catalog %s deleted successfully", result), nil)
}

func (h *Handler) requiredCatalog(c *gin.Context) (string, bool) {
	catalog := c.Param("catalog")
	if catalog == "" {
		utils.ErrorResponse(c, badRequestStatusCode, "catalog name is required", nil)
		return "", false
	}
	return catalog, true
}

func (h *Handler) GetCatalogSpec(c *gin.Context) {
	data, err := os.ReadFile(constants.CatalogSpecFilePath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to read catalog spec", err)
		return
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "invalid catalog spec JSON", err)
		return
	}

	utils.SuccessResponse(c, "catalog spec fetched successfully", dto.SpecResponse{
		Version: constants.CatalogSpecVersion,
		Type:    constants.CatalogTypeIceberg,
		Spec:    spec,
	})
}
