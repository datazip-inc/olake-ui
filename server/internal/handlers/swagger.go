package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

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

	path := strings.TrimPrefix(url, "/swagger")

	// Serve swagger.json and swagger.yaml as static files
	if path == "/swagger.json" || path == "/swagger.yaml" {
		http.ServeFile(h.Ctx.ResponseWriter, h.Ctx.Request, filepath.Join("swagger", filepath.Clean(path)))
		return
	}

	// Serve Swagger UI for all other requests
	httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
	).ServeHTTP(h.Ctx.ResponseWriter, h.Ctx.Request)
}
