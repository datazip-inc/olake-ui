package etl

import (
	"github.com/beego/beego/v2/server/web"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type Handler struct {
	web.Controller
	etl *services.Service
}

var etl *services.Service

func NewHandler(s *services.Service) *Handler {
	etl = s
	return &Handler{etl: s}
}

func (h *Handler) Prepare() {
	h.etl = etl
}
