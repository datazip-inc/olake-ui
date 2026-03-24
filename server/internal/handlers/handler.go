package handlers

import (
	"fmt"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/optimisation"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	optService "github.com/datazip-inc/olake-ui/server/internal/services/optimisation"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

type Handler struct {
	web.Controller
	ETL          *etl.Handler
	Optimisation *optimisation.Handler
}

func NewHandler(appSvc *services.AppService) *Handler {
	h := &Handler{
		ETL: etl.NewHandler(appSvc.ETL()),
	}
	if appSvc.Optimisation() != nil {
		fmt.Println("optimisation is on")
		h.Optimisation = optimisation.NewHandler(appSvc.Optimisation())
	}

	return h
}

func (h *Handler) GetoptimisationStatus() {
	response := map[string]interface{}{
		"enabled": h.Optimisation != nil,
	}

	utils.SuccessResponse(&h.Controller, "optimisation status retrieved successfully", response)
}

func (h *Handler) CreateDestinationAndCatalog() {
	var configJSON string
	if h.Optimisation != nil {
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

	if h.Optimisation != nil && configJSON != "" {
		logger.Debugf("Creating catalog from destination config")
		if _, err := h.Optimisation.GetService().CreateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON); err != nil {
			logger.Errorf("Failed to create catalog: %s", err)
		} else {
			logger.Infof("Catalog created successfully")
		}
	}
}

func (h *Handler) UpdateDestinationAndCatalog() {
	var configJSON string
	if h.Optimisation != nil {
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

	if h.Optimisation != nil && configJSON != "" {
		logger.Debugf("Updating catalog from destination config")
		if _, err := h.Optimisation.GetService().UpdateCatalogFromOLakeConfig(h.Ctx.Request.Context(), configJSON); err != nil {
			logger.Errorf("Failed to update catalog: %s", err)
		} else {
			logger.Infof("Catalog updated successfully")
		}
	}
}

func (h *Handler) DeleteDestinationAndCatalog() {
	var catalogName string
	if h.Optimisation != nil {
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

	if h.Optimisation != nil && catalogName != "" {
		logger.Debugf("Deleting catalog %s", catalogName)
		if _, err := h.Optimisation.GetService().DeleteCatalog(h.Ctx.Request.Context(), catalogName); err != nil {
			logger.Errorf("Failed to delete catalog %s: %s", catalogName, err)
		} else {
			logger.Infof("Catalog %s deleted successfully", catalogName)
		}
	}
}
