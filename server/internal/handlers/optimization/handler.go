package optimization

import (
	"github.com/beego/beego/v2/server/web"
	services "github.com/datazip-inc/olake-ui/server/internal/services/optimization"
)

// encapsulates optimization-specific request
type Handler struct {
	web.Controller
	opt *services.Service
}

// opt holds the singleton service instance injected at startup.
var opt *services.Service

// NewHandler initializes the optimization handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	opt = s
	return &Handler{opt: s}
}

// Prepare runs before each action; Beego constructs a fresh controller per request,
// so we assign the shared service here to avoid nil dereferences.
func (h *Handler) Prepare() {
	h.opt = opt
}
