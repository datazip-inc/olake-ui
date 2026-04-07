package etl

import (
	"github.com/beego/beego/v2/server/web"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

// encapsulates ETL-specific request handling and business logic.
type Handler struct {
	web.Controller
	etl *services.Service
}

// etl holds the singleton service instance injected at startup.
var etl *services.Service

// NewHandler initializes the ETL handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	etl = s
	return &Handler{etl: s}
}

// Prepare runs before each action; Beego constructs a fresh controller per request,
// so we assign the shared AppService here to avoid nil dereferences.
func (h *Handler) Prepare() {
	h.etl = etl
}
