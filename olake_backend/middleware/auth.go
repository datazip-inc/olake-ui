package middleware

import (
	// "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func AuthMiddleware(ctx *context.Context) {
	// Skip auth check for public paths
	if isPublicPath(ctx.Request.URL.Path) {
		return
	}

	// For now, just check Authorization header
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]interface{}{
			"message": "Unauthorized",
			"success": false,
		}, false, false)
		return
	}
}

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/login",
		"/register",
		// Add other public paths
	}
	for _, p := range publicPaths {
		if p == path {
			return true
		}
	}
	return false
} 