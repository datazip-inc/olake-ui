package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type JobHandler struct {
	web.Controller
	jobORM orm.Ormer
}

// Prepare initializes the ORM instance
func (c *JobHandler) Prepare() {
	c.jobORM = orm.NewOrm()
}
