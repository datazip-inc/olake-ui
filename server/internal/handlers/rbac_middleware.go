package handlers

import (
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/datazip-inc/olake-ui/server/internal/auth"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/gin-gonic/gin"
)

func NewRBACMiddleware(enforcer *casbin.Enforcer, sessionsEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !sessionsEnabled || enforcer == nil {
			c.Next()
			return
		}

		userIDVal, exists := c.Get(constants.UserIDContextKey)
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		userID := userIDVal.(int)

		// Global admin bypasses all project-level RBAC checks
		if auth.IsGlobalAdmin(enforcer, userID) {
			c.Next()
			return
		}

		subject := auth.ProjectSubject(userID, c.Param("projectid"))
		allowed, err := enforcer.Enforce(subject, c.Request.URL.Path, c.Request.Method)
		if err != nil || !allowed {
			errorResponse(c, http.StatusForbidden, "forbidden: insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
