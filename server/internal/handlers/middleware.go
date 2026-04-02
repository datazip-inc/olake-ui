package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.sessions.enabled {
			c.Next()
			return
		}

		userID, ok := h.sessions.GetUserID(c)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized, try login again", nil)
			c.Abort()
			return
		}
		c.Set(constants.ContextUserIDKey, userID)
		c.Next()
	}
}
