package handlers

import (
	"net/http"
	"path/filepath"
	"strings"
)

func (h *Handler) ServeSwagger() {
	swaggerDir := "swagger"
	relativePath := strings.TrimPrefix(h.Ctx.Input.URL(), "/swagger")

	if relativePath == "" || relativePath == "/" {
		relativePath = "/index.html"
	}

	// Set Content-Type early
	h.Ctx.Output.ContentType("text/html")

	// Use built-in file serving for efficiency and proper headers
	http.ServeFile(h.Ctx.ResponseWriter, h.Ctx.Request, filepath.Join(swaggerDir, filepath.Clean(relativePath)))
}
