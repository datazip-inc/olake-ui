package controllers // it has api request handlers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type SourceHandler struct {
	web.Controller
	sourceORM orm.Ormer
}

func (c *SourceHandler) Prepare() {
	c.sourceORM = orm.NewOrm()
}
