package main

import (
	_ "olake_backend/routers"
	"olake_backend/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

func init() {
}

func main() {
	utils.ConnectDB()
	if web.BConfig.RunMode == "dev" {
		orm.Debug = true
	}
	web.Run()
}

