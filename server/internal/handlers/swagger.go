package handlers

import (
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// ServeSwagger serves the Swagger UI and static swagger files
func (h *GinHandler) serveSwagger(c *gin.Context) {
	httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	).ServeHTTP(c.Writer, c.Request)
}
