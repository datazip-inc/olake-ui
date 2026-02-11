package handlers

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// ServeSwagger serves the Swagger UI and static swagger files
func (h *Handler) ServeSwagger() {
	url := h.Ctx.Input.URL()

	// Redirect /swagger to /swagger/ for http-swagger compatibility
	if url == "/swagger" {
		h.Redirect("/swagger/", http.StatusMovedPermanently)
		return
	}

	// Serve Swagger UI for all other requests
	httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	).ServeHTTP(h.Ctx.ResponseWriter, h.Ctx.Request)
}
