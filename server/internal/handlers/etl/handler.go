package etl

import (
	"github.com/beego/beego/v2/server/web"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type Handler struct {
	web.Controller
	etl *services.ETLService
}

var etl *services.ETLService

func NewHandler(s *services.ETLService) *Handler {
	etl = s
	return &Handler{etl: s}
}

func (h *Handler) Prepare() {
	h.etl = etl
}
