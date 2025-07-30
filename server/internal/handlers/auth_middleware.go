package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
)

// middleware only works if session is enabled
func AuthMiddleware(ctx *context.Context) {
	if web.BConfig.WebConfig.Session.SessionOn {
		if userID := ctx.Input.Session(constants.SessionUserID); userID == nil {
			// Send unauthorized response
			ctx.Output.SetStatus(401)
			_ = ctx.Output.JSON(models.JSONResponse{
				Message: "Unauthorized, try login again",
				Success: false,
			}, false, false)
		}
	}
}
