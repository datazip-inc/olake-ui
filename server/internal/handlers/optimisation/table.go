package optimisation

import (
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

func (h *Handler) GetTablesWithDetails() {
	catalog, database, ok := h.requiredCatalogAndDatabase()
	if !ok {
		return
	}

	logger.Debugf("Get tables with details initiated catalog[%s] database[%s]", catalog, database)

	tables, err := h.opt.GetTablesWithDetails(h.Ctx.Request.Context(), catalog, database)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get tables with details", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched tables with details", tables)
}

// SetoptimisationCronConfig stores optimisation cron configuration in catalog properties
func (h *Handler) SetProperties() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	var req dto.SQLInput
	if !h.bindJSON(&req) {
		return
	}

	logger.Debugf("Set optimisation cron config initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	result, err := h.opt.SetProperties(h.Ctx.Request.Context(), catalog, database, table, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to set optimisation cron configuration", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

func (h *Handler) requiredCatalogDatabaseTable() (string, string, string, bool) {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog, database, and table parameters are required", nil)
		return "", "", "", false
	}

	return catalog, database, table, true
}

func (h *Handler) requiredCatalogAndDatabase() (string, string, bool) {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")

	if catalog == "" || database == "" {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog and database parameters are required", nil)
		return "", "", false
	}

	return catalog, database, true
}
