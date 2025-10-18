package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/spf13/viper"
)

func (h *Handler) ServeFrontend() {
	indexPath := viper.GetString(constants.FrontendIndexPath)

	// Set Content-Type early
	h.Ctx.Output.ContentType("text/html")

	// Use built-in file serving for efficiency and proper headers
	http.ServeFile(h.Ctx.ResponseWriter, h.Ctx.Request, filepath.Clean(indexPath))
}
