package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httputil"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.sessions.enabled {
			c.Next()
			return
		}
		userID, ok := h.sessions.GetUserID(c)
		if !ok {
			httputil.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized, try login again", nil)
			c.Abort()
			return
		}
		c.Set(httputil.ContextUserIDKey, userID)
		c.Next()
	}
}
