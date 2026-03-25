package optimisation

import (
	"github.com/beego/beego/v2/server/web"
	services "github.com/datazip-inc/olake-ui/server/internal/services/optimisation"
)

// encapsulates optimisation-specific request
type Handler struct {
	web.Controller
	opt *services.Service
}

// opt holds the singleton service instance injected at startup.
var opt *services.Service

// NewHandler initializes the optimisation handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	opt = s
	return &Handler{opt: s}
}

// Prepare runs before each action; Beego constructs a fresh controller per request,
// so we assign the shared service here to avoid nil dereferences.
func (h *Handler) Prepare() {
	h.opt = opt
}

// GetService returns the underlying optimisation service for cross-service orchestration
func (h *Handler) GetService() *services.Service {
	return h.opt
}
