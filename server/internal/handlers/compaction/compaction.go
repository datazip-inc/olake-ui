package compaction

import (
	"net/http"

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
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")

	if catalog == "" || database == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog and database parameters are required", nil)
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

// EnableSelfOptimizing enables self-optimizing for a specific table
func (h *Handler) EnableSelfOptimizing() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	logger.Debugf("Enable self-optimizing initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	result, err := h.compaction.Table.EnableSelfOptimizing(h.Ctx.Request.Context(), catalog, database, table)

	// update catalog table property with format: <db>:<tbl> → <enabled>,<minor>,<major>,<full>
	if err == nil && result.Success {
		// if earlier enabled and then disabled, update the chron config again
		configValue := "true,-1,-1,-1"
		if _, catalogErr := h.compaction.Catalog.SetCatalogTableProperty(h.Ctx.Request.Context(), catalog, database, table, "", configValue); catalogErr != nil {
			logger.Warnf("Failed to update catalog table property for %s.%s.%s: %v", catalog, database, table, catalogErr)
		}
	}

	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to enable self-optimizing", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

func (h *Handler) DisableSelfOptimizing() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	logger.Debugf("Disable self-optimizing initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	result, err := h.compaction.Table.DisableSelfOptimizing(h.Ctx.Request.Context(), catalog, database, table)

	// update catalog table property with format: <db>:<tbl> → <enabled>,<minor>,<major>,<full>
	if err == nil && result.Success {
		configValue := "false,-1,-1,-1"
		if _, catalogErr := h.compaction.Catalog.SetCatalogTableProperty(h.Ctx.Request.Context(), catalog, database, table, "", configValue); catalogErr != nil {
			logger.Warnf("Failed to update catalog table property for %s.%s.%s: %v", catalog, database, table, catalogErr)
		}
	}

	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to disable self-optimizing", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// fetches detailed file metrics for a specific table
func (h *Handler) GetTableMetrics() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "catalog, database, and table parameters are required", nil)
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
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	var config models.CompactionCronConfigRequest
	if err := h.Ctx.BindJSON(&config); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	logger.Debugf("Set compaction cron config initiated catalog[%s] database[%s] table[%s]", catalog, database, table)

	result, err := h.compaction.Optimization.SetCompactionCronConfig(h.Ctx.Request.Context(), catalog, database, table, config)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to set compaction cron configuration", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// GetCompactionCronConfig retrieves compaction cron configuration from catalog properties
func (h *Handler) GetCompactionCronConfig() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
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
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	// pagination
	page, _ := h.GetInt("page", 1)
	pageSize, _ := h.GetInt("pageSize", 1000)

	logger.Debugf("Get compaction runs initiated catalog[%s] database[%s] table[%s] page[%d] pageSize[%d]", catalog, database, table, page, pageSize)

	runs, err := h.compaction.Optimization.GetCompactionRuns(h.Ctx.Request.Context(), catalog, database, table, page, pageSize)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get compaction runs", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched compaction runs", runs)
}

func (h *Handler) GetCatalog() {
	catalogName := h.Ctx.Input.Param(":catalog")
	if catalogName == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "catalog name is required", nil)
		return
	}

	logger.Debugf("Get catalog details initiated catalog[%s]", catalogName)

	catalog, err := h.compaction.Catalog.GetCatalog(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to get catalog details", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "successfully fetched catalog details", catalog)
}

func (h *Handler) CreateCatalog() {
	var req models.CatalogRequest
	if err := h.Ctx.BindJSON(&req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "invalid request body", err)
		return
	}

	logger.Debugf("Create catalog initiated name[%s]", req.Name)

	result, err := h.compaction.Catalog.CreateCatalog(h.Ctx.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to create catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, nil)
}

// updates an existing catalog
func (h *Handler) UpdateCatalog() {
	catalogName := h.Ctx.Input.Param(":catalog")
	if catalogName == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "catalog name is required", nil)
		return
	}

	var req models.CatalogRequest
	if err := h.Ctx.BindJSON(&req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	logger.Debugf("Update catalog initiated catalog[%s]", catalogName)

	result, err := h.compaction.Catalog.UpdateCatalog(h.Ctx.Request.Context(), catalogName, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to update catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, nil)
}

// deletes a catalog
func (h *Handler) DeleteCatalog() {
	catalogName := h.Ctx.Input.Param(":catalog")
	if catalogName == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog name is required", nil)
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

// cancels a running compaction process
func (h *Handler) CancelCompactionProcess() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")
	processID := h.Ctx.Input.Param(":processid")

	if catalog == "" || database == "" || table == "" || processID == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, table, and process ID are required", nil)
		return
	}

	logger.Debugf("Cancel compaction process initiated catalog[%s] database[%s] table[%s] processID[%s]", catalog, database, table, processID)

	err := h.compaction.Optimization.CancelCompactionProcess(h.Ctx.Request.Context(), catalog, database, table, processID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to cancel compaction process", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully canceled compaction process", map[string]string{
		"catalog":    catalog,
		"database":   database,
		"table":      table,
		"process_id": processID,
	})
}
