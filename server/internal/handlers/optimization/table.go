package optimization

import (
	"strconv"

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

// GetTableRuns lists compaction and snapshot-expiry runs for a table.
// Query: type=all|expiry|optimization (default all), status, page, pageSize.
func (h *Handler) GetTableRuns(c *gin.Context) {
	catalog, database, ok := h.requiredCatalogAndDatabase(c)
	if !ok {
		return
	}
	table := c.Param("table")
	if table == "" {
		utils.ErrorResponse(c, badRequestStatusCode, "table parameter is required", nil)
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		utils.ErrorResponse(c, badRequestStatusCode, "page must be a positive integer", err)
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 {
		utils.ErrorResponse(c, badRequestStatusCode, "pageSize must be a positive integer", err)
		return
	}

	runs, err := h.opt.GetTableRuns(c.Request.Context(), catalog, database, table, c.Query("type"), c.Query("status"), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(c, "Successfully fetched table runs", runs)
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
