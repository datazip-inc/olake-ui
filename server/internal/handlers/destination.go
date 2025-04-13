package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DestHandler struct {
	web.Controller
	destORM orm.Ormer
}

func (c *DestHandler) Prepare() {
	c.destORM = orm.NewOrm()
}
