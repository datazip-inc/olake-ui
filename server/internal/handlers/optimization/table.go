package optimization

import (
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
)

func (h *Handler) GetTablesWithDetails() {
	catalog, database, ok := h.requiredCatalogAndDatabase()
	if !ok {
		return
	}

	tables, err := h.opt.GetTablesWithDetails(h.Ctx.Request.Context(), catalog, database)
	if err != nil {
		utils.ErrorResponse(&h.Controller, upstreamStatus(err), err.Error(), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched tables with details", tables)
}

// configures table level properties for optimization
func (h *Handler) SetProperties() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	var req dto.SQLInput
	if err := h.bindJSON(&req); err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid request body", err)
		return
	}

	result, err := h.opt.SetProperties(h.Ctx.Request.Context(), catalog, database, table, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, upstreamStatus(err), err.Error(), err)
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
