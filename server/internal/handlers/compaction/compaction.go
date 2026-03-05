package compaction

import (
	"net/http"
	"os"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

func getCompactionClient() *compaction.Compaction {
	baseURL := os.Getenv("AMORO_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:1630"
	}

	apiKey := os.Getenv("AMORO_API_KEY")
	apiSecret := os.Getenv("AMORO_API_SECRET")

	return compaction.NewClient(baseURL, apiKey, apiSecret)
}

// func (h *Handler) GetDashboard() {
// 	client := getCompactionClient()
// 	dashboard, err := client.GetDashboard(h.Ctx.Request.Context())
// 	if err != nil {
// 		logger.Errorf("Failed to get dashboard data: %v", err)
// 		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get dashboard data", err)
// 		return
// 	}

// 	utils.SuccessResponse(&h.Controller, "Successfully fetched dashboard data", dashboard)
// }

func (h *Handler) GetCatalogsWithDatabases() {
	client := getCompactionClient()
	catalogs, err := client.GetCatalogsWithDatabases(h.Ctx.Request.Context())
	if err != nil {
		logger.Errorf("Failed to get catalogs with databases: %v", err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get catalogs with databases", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched catalogs with databases", catalogs)
}

func (h *Handler) GetTablesWithDetails() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")

	if catalog == "" || database == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog and database parameters are required", nil)
		return
	}

	client := getCompactionClient()
	tables, err := client.GetTablesWithDetails(h.Ctx.Request.Context(), catalog, database, h.db)
	if err != nil {
		logger.Errorf("Failed to get tables with details for catalog %s, database %s: %v", catalog, database, err)
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

	client := getCompactionClient()
	result, err := client.EnableSelfOptimizing(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		logger.Errorf("Failed to enable self-optimizing for %s.%s.%s: %v", catalog, database, table, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to enable self-optimizing", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// DisableSelfOptimizing disables self-optimizing for a specific table
func (h *Handler) DisableSelfOptimizing() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	client := getCompactionClient()
	result, err := client.DisableSelfOptimizing(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		logger.Errorf("Failed to disable self-optimizing for %s.%s.%s: %v", catalog, database, table, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to disable self-optimizing", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// SetTableProperties sets custom properties for a table
func (h *Handler) SetTableProperties() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	var req compaction.SetTablePropertiesRequest
	if err := h.Ctx.BindJSON(&req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Override catalog, database, table from URL params
	req.Catalog = catalog
	req.Database = database
	req.Table = table

	client := getCompactionClient()
	result, err := client.SetTableProperties(h.Ctx.Request.Context(), req)
	if err != nil {
		logger.Errorf("Failed to set table properties for %s.%s.%s: %v", catalog, database, table, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to set table properties", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// GetTableMetrics fetches detailed file metrics for a specific table
func (h *Handler) GetTableMetrics() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	client := getCompactionClient()
	metrics, err := client.GetTableMetrics(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		logger.Errorf("Failed to get table metrics for %s.%s.%s: %v", catalog, database, table, err)
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

	var config compaction.CompactionCronConfigRequest
	if err := h.Ctx.BindJSON(&config); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	client := getCompactionClient()
	result, err := client.SetCompactionCronConfig(h.Ctx.Request.Context(), catalog, database, table, config)
	if err != nil {
		logger.Errorf("Failed to set compaction cron config for %s.%s.%s: %v", catalog, database, table, err)
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

	client := getCompactionClient()
	config, err := client.GetCompactionCronConfig(h.Ctx.Request.Context(), catalog, database, table)
	if err != nil {
		logger.Errorf("Failed to get compaction cron config for %s.%s.%s: %v", catalog, database, table, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get compaction cron configuration", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched compaction cron configuration", config)
}

// GetCompactionRuns fetches the list of compaction runs/processes for a table
func (h *Handler) GetCompactionRuns() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")

	if catalog == "" || database == "" || table == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, and table parameters are required", nil)
		return
	}

	// Get pagination parameters
	page, _ := h.GetInt("page", 1)
	if page < 1 {
		page = 1
	}

	pageSize, _ := h.GetInt("pageSize", 20)
	if pageSize < 1 {
		pageSize = 20
	}

	client := getCompactionClient()
	runs, err := client.GetCompactionRuns(h.Ctx.Request.Context(), catalog, database, table, page, pageSize)
	if err != nil {
		logger.Errorf("Failed to get compaction runs for %s.%s.%s: %v", catalog, database, table, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get compaction runs", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched compaction runs", runs)
}

// GetProcessMetrics fetches the metrics for a specific compaction process/run
func (h *Handler) GetProcessMetrics() {
	catalog := h.Ctx.Input.Param(":catalog")
	database := h.Ctx.Input.Param(":database")
	table := h.Ctx.Input.Param(":table")
	runID := h.Ctx.Input.Param(":runid")

	if catalog == "" || database == "" || table == "" || runID == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog, database, table, and runid parameters are required", nil)
		return
	}

	client := getCompactionClient()
	metrics, err := client.GetProcessMetrics(h.Ctx.Request.Context(), catalog, database, table, runID)
	if err != nil {
		logger.Errorf("Failed to get process metrics for %s.%s.%s process %s: %v", catalog, database, table, runID, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get process metrics", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "Successfully fetched process metrics", metrics)
}

// CreateCatalog creates a new catalog
func (h *Handler) CreateCatalog() {
	var req compaction.CatalogRequest
	if err := h.Ctx.BindJSON(&req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	client := getCompactionClient()
	result, err := client.CreateCatalog(h.Ctx.Request.Context(), req)
	if err != nil {
		logger.Errorf("Failed to create catalog %s: %v", req.Name, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to create catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// UpdateCatalog updates an existing catalog
func (h *Handler) UpdateCatalog() {
	catalogName := h.Ctx.Input.Param(":catalog")
	if catalogName == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog name is required", nil)
		return
	}

	var req compaction.CatalogRequest
	if err := h.Ctx.BindJSON(&req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	client := getCompactionClient()
	result, err := client.UpdateCatalog(h.Ctx.Request.Context(), catalogName, req)
	if err != nil {
		logger.Errorf("Failed to update catalog %s: %v", catalogName, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to update catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}

// DeleteCatalog deletes a catalog
func (h *Handler) DeleteCatalog() {
	catalogName := h.Ctx.Input.Param(":catalog")
	if catalogName == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Catalog name is required", nil)
		return
	}

	client := getCompactionClient()
	result, err := client.DeleteCatalog(h.Ctx.Request.Context(), catalogName)
	if err != nil {
		logger.Errorf("Failed to delete catalog %s: %v", catalogName, err)
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to delete catalog", err)
		return
	}

	utils.SuccessResponse(&h.Controller, result.Message, result)
}
