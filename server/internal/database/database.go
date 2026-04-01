package database

import (
	"fmt"
	"net/url"

	_ "github.com/lib/pq" // required for registering driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

type Database struct {
	conn *gorm.DB
}

func Init() (*Database, error) {
	cfg := appconfig.Load()

	uri, err := BuildPostgresURIFromConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build postgres uri: %s", err)
	}

	logLevel := gormlogger.Warn
	if cfg.RunMode == "dev" || cfg.RunMode == "localdev" {
		logLevel = gormlogger.Info
	}

	conn, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %s", err)
	}

	// migration for database tables
	if err := conn.AutoMigrate(
		new(models.ProjectSettings),
		new(models.Source),
		new(models.Destination),
		new(models.Job),
		new(models.User),
		new(models.Catalog),
		new(models.ProjectUserRole),
	); err != nil {
		return nil, fmt.Errorf("failed to run automigrate: %s", err)
	}

	// Add session table if sessions are enabled
	if cfg.SessionOn {
		err = conn.Exec(`CREATE TABLE IF NOT EXISTS session (
			session_key VARCHAR(64) PRIMARY KEY,
			session_data BYTEA,
			session_expiry TIMESTAMP WITH TIME ZONE);`).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create session table: %s", err)
		}
	}
	return &Database{conn: conn}, nil
}

// BuildPostgresURIFromConfig reads POSTGRES_DB_HOST, POSTGRES_DB_PORT, etc. from app.conf
// and constructs the Postgres connection URI.
func BuildPostgresURIFromConfig() (string, error) {
	logger.Info("Building Postgres URI from config")
	cfg := appconfig.Load()

	// First, check if postgresdb is set directly
	if cfg.PostgresDSN != "" {
		return cfg.PostgresDSN, nil
	}

	user := cfg.OlakePostgresUser
	password := cfg.OlakePostgresPassword
	host := cfg.OlakePostgresHost
	port := cfg.OlakePostgresPort
	dbName := cfg.OlakePostgresDBName
	sslMode := cfg.OlakePostgresSSLMode

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
func (d *Database) Conn() *gorm.DB {
	return d.conn
}
