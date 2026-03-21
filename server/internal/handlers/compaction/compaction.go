package compaction

import (
	"net/http"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// get the catalogs along with the databases in those catalogs
func (h *Handler) GetCatalogsWithDatabases() {
	logger.Debugf("Get catalogs with databases initiated")

	catalogs, err := h.compaction.Aggregator.GetCatalogsWithDatabases(h.Ctx.Request.Context())
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to get catalogs with databases", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "successfully fetched catalogs with databases", catalogs)
}

func (h *Handler) GetTablesWithDetails() {
	catalog, database, ok := h.requiredCatalogAndDatabase()
	if !ok {
		return
	}

	logger.Debugf("Get tables with details initiated catalog[%s] database[%s]", catalog, database)

	tables, err := h.compaction.Aggregator.GetTablesWithDetails(h.Ctx.Request.Context(), catalog, database, h.db)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get tables with details", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched tables with details", tables)
}

// EnableSelfOptimizing enables self-optimizing for a table.
func (h *Handler) EnableSelfOptimizing() {
	h.toggleSelfOptimizing(true)
}

// DisableSelfOptimizing disables self-optimizing for a table.
func (h *Handler) DisableSelfOptimizing() {
	h.toggleSelfOptimizing(false)
}

func (h *Handler) toggleSelfOptimizing(enable bool) {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	action := "disable"
	service := h.compaction.Table.DisableSelfOptimizing
	errorMessage := "failed to disable self-optimizing"

	if enable {
		action = "enable"
		service = h.compaction.Table.EnableSelfOptimizing
		errorMessage = "failed to enable self-optimizing"
	}

	logger.Debugf("%s self-optimizing initiated catalog[%s] database[%s] table[%s]", strings.Title(action), catalog, database, table)

	ctx := h.Ctx.Request.Context()
	result, err := service(ctx, catalog, database, table)
	if err != nil {
		utils.ErrorResponse(&h.Controller, internalServerErrorStatusCode, errorMessage, err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// fetches detailed file metrics for a specific table
func (h *Handler) GetTableMetrics() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	logger.Debugf("Get table metrics initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	metrics, err := h.compaction.Optimization.GetTableMetrics(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get table metrics", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched table metrics", metrics)
}

// SetCompactionCronConfig stores compaction cron configuration in catalog properties
func (h *Handler) SetCompactionCronConfig() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	var req models.CompactionCronConfigRequest
	if !h.bindJSON(&req) {
		return
	}

	logger.Debugf("Set compaction cron config initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	result, err := h.compaction.Optimization.SetCompactionCronConfig(h.Ctx.Request.Context(), catalog, database, table, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to set compaction cron configuration", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// GetCompactionCronConfig retrieves compaction cron configuration from catalog properties
func (h *Handler) GetCompactionCronConfig() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	logger.Debugf("Get compaction cron config initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	config, err := h.compaction.Optimization.GetCompactionCronConfig(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get compaction cron configuration", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched compaction cron configuration", config)
}

// fetches the list of compaction runs/processes for a particular table
func (h *Handler) GetCompactionRuns() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	page, pageSize := h.pagination(defaultCompactionRunsPage, defaultCompactionRunsPageSize)

	logger.Debugf("Get compaction runs initiated catalog[%s] database[%s] table[%s] page[%d] pageSize[%d]", catalog, database, table, page, pageSize)

	runs, err := h.compaction.Optimization.GetCompactionRuns(h.Ctx.Request.Context(), catalog, database, table, page, pageSize)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get compaction runs", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched compaction runs", runs)
}

func (h *Handler) GetCatalog() {
	catalogName, ok := h.requiredCatalog()
	if !ok {
		return
	}

	logger.Debugf("Get catalog details initiated catalog[%s]", catalogName)

	// Get catalog in Olake config format
	olakeConfig, err := h.compaction.Catalog.GetCatalogAsOLakeConfig(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to get catalog details", err)
		return
	}

	// Return only the config data without wrapper
	h.Controller.Ctx.Output.SetStatus(http.StatusOK)
	h.Controller.Data["json"] = olakeConfig
	_ = h.Controller.ServeJSON()
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

	result, err := h.compaction.Catalog.CreateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON)
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

	result, err := h.compaction.Catalog.UpdateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON)
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

	result, err := h.compaction.Catalog.DeleteCatalog(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to delete catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, nil)
}

// cancels a running compaction process by fetching the latest process and validating its status
func (h *Handler) CancelCompactionProcess() {
	catalog, database, table, ok := h.requiredCatalogDatabaseTable()
	if !ok {
		return
	}

	logger.Debugf("Cancel compaction process initiated for catalog[%s] database[%s] table[%s]", catalog, database, table)

	processID, err := h.compaction.Optimization.CancelLatestCompactionProcess(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Failed to cancel compaction process", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully canceled compaction process", map[string]string{
		"catalog":    catalog,
		"database":   database,
		"table":      table,
		"process_id": processID,
	})
}
