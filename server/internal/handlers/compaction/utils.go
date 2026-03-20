package compaction

import (
	"net/http"
	"strings"

	"github.com/datazip-inc/olake-ui/server/utils"
)

const (
	defaultCompactionRunsPage     = 1
	defaultCompactionRunsPageSize = 1000
	selfOptimizingEnabledConfig   = "true,-1,-1,-1"
	selfOptimizingDisabledConfig  = "false,-1,-1,-1"
	internalServerErrorStatusCode = http.StatusInternalServerError
	badRequestStatusCode          = http.StatusBadRequest
)

func (h *Handler) requiredCatalog() (string, bool) {
	catalog := h.Ctx.Input.Param(":catalog")
	if catalog == "" {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "catalog name is required", nil)
		return "", false
	}
	return catalog, true
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

func (h *Handler) bindJSON(dst interface{}) bool {
	if err := h.Ctx.BindJSON(dst); err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid request body", err)
		return false
	}
	return true
}

func (h *Handler) param(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(h.Ctx.Input.Param(name)); value != "" {
			return value
		}
	}
	return ""
}

func (h *Handler) pagination(defaultPage, defaultPageSize int) (int, int) {
	page, err := h.GetInt("page", defaultPage)
	if err != nil || page < 1 {
		page = defaultPage
	}

	pageSize, err := h.GetInt("pageSize", defaultPageSize)
	if err != nil || pageSize < 1 {
		pageSize = defaultPageSize
	}

	return page, pageSize
}
