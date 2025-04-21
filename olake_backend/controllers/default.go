package controllers

import (
	"github.com/beego/beego/v2/server/web"
)

type MainController struct {
	web.Controller
}

func (c *MainController) Get() {
	c.Data["json"] = map[string]string{"message": "hello world"}
	c.ServeJSON()
}
