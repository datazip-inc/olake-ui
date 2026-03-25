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

// NewHandler initializes the optimisation handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	return &Handler{opt: s}
}

// GetService returns the underlying optimisation service for cross-service orchestration
func (h *Handler) GetService() *services.Service {
	return h.opt
}
