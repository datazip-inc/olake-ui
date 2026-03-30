package optimization

import (
	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httputil"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

func (h *Handler) GetTablesWithDetails(c *gin.Context) {
	catalog, database, ok := h.requiredCatalogAndDatabase(c)
	if !ok {
		return
	}

	tables, err := h.opt.GetTablesWithDetails(c.Request.Context(), catalog, database)
	if err != nil {
		httputil.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	httputil.SuccessResponse(c, "Successfully fetched tables with details", tables)
}

// SetoptimizationCronConfig stores optimization cron configuration in catalog properties
func (h *Handler) SetProperties(c *gin.Context) {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable(c)
	if !ok {
		return
	}

	var req dto.SQLInput
	if !h.bindJSON(c, &req) {
		return
	}

	result, err := h.opt.SetProperties(c.Request.Context(), catalog, database, table, req)
	if err != nil {
		httputil.ErrorResponse(c, upstreamStatus(err), err.Error(), err)
		return
	}

	httputil.SuccessResponse(c, result.Message, result)
}

func (h *Handler) requiredCatalogDatabaseTable(c *gin.Context) (string, string, string, bool) {
	catalog := c.Param("catalog")
	database := c.Param("database")
	table := c.Param("table")

	if catalog == "" || database == "" || table == "" {
		httputil.ErrorResponse(c, badRequestStatusCode, "catalog, database, and table parameters are required", nil)
		return "", "", "", false
	}

	return catalog, database, table, true
}

func (h *Handler) requiredCatalogAndDatabase(c *gin.Context) (string, string, bool) {
	catalog := c.Param("catalog")
	database := c.Param("database")

	if catalog == "" || database == "" {
		httputil.ErrorResponse(c, badRequestStatusCode, "catalog and database parameters are required", nil)
		return "", "", false
	}

	return catalog, database, true
}
