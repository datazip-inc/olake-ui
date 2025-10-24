package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/etl-service"
)

type Handler struct {
	web.Controller
	etl *services.ETLService
}

// appService holds the singleton service instance injected at startup.
var etl *services.ETLService

func NewHandler(s *services.ETLService) *Handler {
	etl = s
	return &Handler{etl: s}
}

// Prepare runs before each action; Beego constructs a fresh controller per request,
// so we assign the shared AppService here to avoid nil dereferences.
func (h *Handler) Prepare() {
	h.etl = etl
}
