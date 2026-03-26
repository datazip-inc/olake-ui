package handlers

import (
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"

	optService "github.com/datazip-inc/olake-ui/server/internal/services/optimization"
)

// CreateDestinationAndCatalog orchestrates ETL destination creation with optional catalog creation.
// This cross-cutting concern requires access to both ETL and optimization handlers.
func (h *Handler) CreateDestinationAndCatalog() {
	var configJSON string
	if h.Optimization != nil {
		var req dto.CreateDestinationRequest
		if err := dto.UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err == nil {
			configJSON = req.Config
		}
	}

	h.ETL.Controller = h.Controller
	h.ETL.CreateDestination()

	if h.Ctx.ResponseWriter.Status >= 400 {
		return
	}

	if h.Optimization != nil && configJSON != "" {
		catalogName, _ := optService.ExtractCatalogNameFromConfig(configJSON)
		logger.Debugf("Creating catalog[%s] from destination config", catalogName)

		if _, err := h.Optimization.GetService().CreateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON, true); err != nil {
			logger.Errorf("Failed to create catalog[%s]: %s", catalogName, err)
			// Return warning in response - destination created but catalog failed
			utils.ErrorResponse(&h.Controller, http.StatusPartialContent, "destination created but catalog creation failed", err)
			return
		}
		logger.Infof("Catalog[%s] created successfully", catalogName)
	}
}

// UpdateDestinationAndCatalog orchestrates ETL destination updates with catalog synchronization.
// This ensures catalog configs stay in sync when destinations are modified.
func (h *Handler) UpdateDestinationAndCatalog() {
	var configJSON string
	if h.Optimization != nil {
		var req dto.UpdateDestinationRequest
		if err := dto.UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err == nil {
			configJSON = req.Config
		}
	}

	h.ETL.Controller = h.Controller
	h.ETL.UpdateDestination()

	if h.Ctx.ResponseWriter.Status >= 400 {
		return
	}

	if h.Optimization != nil && configJSON != "" {
		catalogName, _ := optService.ExtractCatalogNameFromConfig(configJSON)
		logger.Debugf("Updating catalog[%s] from destination config", catalogName)

		if _, err := h.Optimization.GetService().UpdateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON); err != nil {
			logger.Errorf("Failed to update catalog[%s]: %s", catalogName, err)
			// Return warning - destination updated but catalog sync failed
			utils.ErrorResponse(&h.Controller, http.StatusPartialContent, "destination updated but catalog sync failed", err)
			return
		}
		logger.Infof("Catalog[%s] updated successfully", catalogName)
	}
}

// DeleteDestinationAndCatalog orchestrates ETL destination deletion with catalog cleanup.
// This prevents orphaned catalogs when their corresponding destinations are removed.
func (h *Handler) DeleteDestinationAndCatalog() {
	var catalogName string
	if h.Optimization != nil {
		id, err := etl.GetIDFromPath(&h.Controller)
		if err == nil {
			if destination, err := h.ETL.GetService().GetDestinationByID(id); err == nil {
				catalogName, _ = optService.ExtractCatalogNameFromConfig(destination.Config)
			}
		}
	}

	h.ETL.Controller = h.Controller
	h.ETL.DeleteDestination()

	if h.Ctx.ResponseWriter.Status >= 400 {
		return
	}

	if h.Optimization != nil && catalogName != "" {
		logger.Debugf("Deleting catalog[%s]", catalogName)

		if _, err := h.Optimization.GetService().DeleteCatalog(h.Ctx.Request.Context(), catalogName); err != nil {
			logger.Errorf("Failed to delete catalog[%s]: %s", catalogName, err)
			// Log warning but don't fail the entire operation - destination already deleted
			logger.Warnf("Destination deleted successfully but catalog[%s] cleanup failed - may need manual cleanup", catalogName)
		} else {
			logger.Infof("Catalog[%s] deleted successfully", catalogName)
		}
	}
}
