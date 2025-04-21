package utils

import "github.com/beego/beego/v2/server/web"

type ErrorResponse struct {
    Message string `json:"message"`
    Code    int    `json:"code"`
}

func HandleError(c *web.Controller, status int, message string) {
    c.Ctx.Output.SetStatus(status)
    c.Data["json"] = ErrorResponse{
        Message: message,
        Code:    status,
    }
    c.ServeJSON()
} 