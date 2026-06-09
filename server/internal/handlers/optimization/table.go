package optimization

import (
	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

func (h *Handler) GetTablesWithDetails(c *gin.Context) {
	catalog, database, ok := h.requiredCatalogAndDatabase(c)
	if !ok {
		return
	}

	tables, err := h.opt.GetTablesWithDetails(c.Request.Context(), catalog, database)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, "Successfully fetched tables with details", tables)
}

// SetProperties configures the same optimization properties on multiple tables (one terminal batch).
func (h *Handler) SetProperties(c *gin.Context) {
	catalog, database, ok := h.requiredCatalogAndDatabase(c)
	if !ok {
		return
	}

	var tableConfigs dto.OptimizationTableConfig
	if err := utils.BindAndValidate(c, &tableConfigs); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), "invalid request body for optimization table config", err)
		return
	}

	result, err := h.opt.SetProperties(c.Request.Context(), catalog, database, tableConfigs)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, "Finished setting properties for selected tables", result)
}

func (h *Handler) requiredCatalogAndDatabase(c *gin.Context) (string, string, bool) {
	catalog := c.Param("catalog")
	database := c.Param("database")

	if catalog == "" || database == "" {
		utils.ErrorResponse(c, badRequestStatusCode, "catalog and database parameters are required", nil)
		return "", "", false
	}

	return catalog, database, true
}
