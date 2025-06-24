package database

import (
	"encoding/gob"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/beego/beego/v2/server/web/session/postgres" // required for session
	_ "github.com/lib/pq"                                     // required for registering driver

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
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
		web.BConfig.WebConfig.Session.SessionCookieLifeTime = 30 * 24 * 60 * 60 // 30 days
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

	// Create tables if they do not exist
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		return fmt.Errorf("failed to sync database schema: %s", err)
	}
	// Add session table if sessions are enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_, err = orm.NewOrm().Raw(`CREATE TABLE IF NOT EXISTS session (
    session_key VARCHAR(64) PRIMARY KEY,
    session_data BYTEA,
    session_expiry TIMESTAMP WITH TIME ZONE
);`).Exec()

		if err != nil {
			return fmt.Errorf("failed to create session table: %s", err)
		}
	}
	return nil
}
