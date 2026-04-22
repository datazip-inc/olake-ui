package optimization

import (
	"errors"
	"net/http"

	services "github.com/datazip-inc/olake-ui/server/internal/services/optimization"
)

// encapsulates optimization-specific request
type Handler struct {
	opt *services.Service
}

// NewHandler initializes the optimization handler with its service dependency.
func NewHandler(s *services.Service) *Handler {
	return &Handler{opt: s}
}

// returns the HTTP status code carried by an upstream optimzation
func upstreamStatus(err error) int {
	var httpErr *services.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode
	}
	return http.StatusInternalServerError
}
