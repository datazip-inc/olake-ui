package database

import (
	"encoding/gob"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/beego/beego/v2/server/web/session/postgres" // required for session
	_ "github.com/lib/pq"                                     // required for registering driver

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/models"
)

func Init(uri string) error {
	// register driver
	err := orm.RegisterDriver("postgres", orm.DRPostgres)
	if err != nil {
		return fmt.Errorf("failed to register postgres driver: %s", err)
	}

	// register database
	err = orm.RegisterDataBase("default", "postgres", uri)
	if err != nil {
		return fmt.Errorf("failed to register postgres database: %s", err)
	}

	// enable session by default
	if web.BConfig.WebConfig.Session.SessionOn {
		web.BConfig.WebConfig.Session.SessionName = "olake-session"
		web.BConfig.WebConfig.Session.SessionProvider = "postgresql"
		web.BConfig.WebConfig.Session.SessionProviderConfig = uri
		web.BConfig.WebConfig.Session.SessionCookieLifeTime = 10800 // 3 hour
	}

	// register session user
	gob.Register(constants.SessionUserID)
	// register models in order of dependency or foreign key constraints
	orm.RegisterModel(
		new(models.Source),
		new(models.Destination),
		new(models.Job),
		new(models.User),
		new(models.Catalog),
	)
	if web.BConfig.WebConfig.Session.SessionOn {
		orm.RegisterModel(new(models.Session))
	}

	// Create tables if they do not exist
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		return fmt.Errorf("failed to sync database schema: %s", err)
	}
	return nil
}
