package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/beego/beego/v2/server/web"
)

type FrontendHandler struct {
	web.Controller
}

func (c *FrontendHandler) Get() {
	const indexPath = "/opt/frontend/dist/index.html"

	// Set Content-Type early
	c.Ctx.Output.ContentType("text/html")

	// Use built-in file serving for efficiency and proper headers
	http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, filepath.Clean(indexPath))
}
