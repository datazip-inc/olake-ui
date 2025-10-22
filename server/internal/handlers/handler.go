package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/services"
)

type Handler struct {
	web.Controller
	svc *services.AppService
}

// appService holds the singleton service instance injected at startup.
var appService *services.AppService

// NewHandler registers the shared AppService for later injection into each request-scoped handler.
func NewHandler(s *services.AppService) *Handler {
	appService = s
	return &Handler{}
}

// Prepare runs before each action; Beego constructs a fresh controller per request,
// so we assign the shared AppService here to avoid nil dereferences.
func (h *Handler) Prepare() {
	h.svc = appService
}
