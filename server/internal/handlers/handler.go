package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	"github.com/datazip-inc/olake-ui/server/internal/services"
)

type Handler struct {
	web.Controller
	ETL *etl.Handler
}

func NewHandler(appSvc *services.AppService) *Handler {
	return &Handler{
		ETL: etl.NewHandler(appSvc.ETL()),
	}
}
