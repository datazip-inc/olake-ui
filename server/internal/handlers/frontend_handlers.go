package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/server/web"
)

type FrontendHandler struct {
	web.Controller
}

func (c *FrontendHandler) Get() {
	// Get isDev from environment variable, default to false if not set
	isDev := strings.ToLower(os.Getenv("IS_DEV")) == "true"

	if isDev {
		// In development mode, proxy to Vite dev server
		viteDevServer, err := url.Parse("http://localhost:5173")
		if err != nil {
			c.Ctx.Output.SetStatus(500)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(viteDevServer)
		
		// Add error handling to the proxy
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(fmt.Sprintf("Proxy error: %v", err)))
		}

		// Update the Host header to match Vite dev server
		c.Ctx.Request.Host = viteDevServer.Host
		proxy.ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
		return
	}

	// Production mode - serve from dist
	const indexPath = "/opt/frontend/dist/index.html"
	// Set Content-Type early
	c.Ctx.Output.ContentType("text/html")

	// Use built-in file serving for efficiency and proper headers
	http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, filepath.Clean(indexPath))
}
