package database

import (
	"encoding/gob"
	"fmt"
	"net/url"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/beego/beego/v2/server/web/session/postgres" // required for session
	_ "github.com/lib/pq"                                     // required for registering driver

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

type Database struct {
	ormer orm.Ormer
}

func Init() (*Database, error) {
	// register driver
	uri, err := BuildPostgresURIFromConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build postgres uri: %s", err)
	}

	err = orm.RegisterDriver("postgres", orm.DRPostgres)
	if err != nil {
		return nil, fmt.Errorf("failed to register postgres driver: %s", err)
	}

	// register database
	err = orm.RegisterDataBase("default", "postgres", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to register postgres database: %s", err)
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
		new(models.ProjectSettings),
		new(models.Source),
		new(models.Destination),
		new(models.Job),
		new(models.User),
		new(models.Catalog),
	)

	// Create tables if they do not exist
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sync database schema: %s", err)
	}

	// Add session table if sessions are enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_, err = orm.NewOrm().Raw(`CREATE TABLE IF NOT EXISTS session (
    session_key VARCHAR(64) PRIMARY KEY,
    session_data BYTEA,
    session_expiry TIMESTAMP WITH TIME ZONE
);`).Exec()

		if err != nil {
			return nil, fmt.Errorf("failed to create session table: %s", err)
		}
	}
	return &Database{ormer: orm.NewOrm()}, nil
}

// BuildPostgresURIFromConfig reads POSTGRES_DB_HOST, POSTGRES_DB_PORT, etc. from app.conf
// and constructs the Postgres connection URI.
func BuildPostgresURIFromConfig() (string, error) {
	logger.Info("Building Postgres URI from config")

	// First, check if postgresdb is set directly
	if dsn, err := web.AppConfig.String(constants.ConfPostgresDB); err == nil && dsn != "" {
		return dsn, nil
	}

	user, _ := web.AppConfig.String(constants.ConfOLakePostgresUser)
	password, _ := web.AppConfig.String(constants.ConfOLakePostgresPassword)
	host, _ := web.AppConfig.String(constants.ConfOLakePostgresHost)
	port, _ := web.AppConfig.String(constants.ConfOLakePostgresPort)
	dbName, _ := web.AppConfig.String(constants.ConfOLakePostgresDBname)
	sslMode, _ := web.AppConfig.String(constants.ConfOLakePostgresSslmode)

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + url.PathEscape(dbName),
	}

	query := u.Query()
	query.Set("sslmode", sslMode)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
