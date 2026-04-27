package database

import (
	"fmt"
	"net/url"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
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
	if cfg.RunMode == "localdev" {
		logLevel = gormlogger.Info
	}

	conn, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %s", err)
	}

	// Clean up constraints left by beego-orm and early GORM versions before
	// AutoMigrate runs, so upgrades from <= 0.4.1 don't fail on duplicates.
	if err := dropLegacyConstraints(conn); err != nil {
		return nil, fmt.Errorf("failed to drop legacy constraints: %s", err)
	}

	// Order matters for fresh installs: tables with no FK dependencies first,
	// then tables that reference them. Source/Destination/Job all FK → User.
	if err := conn.AutoMigrate(
		new(models.User),
		new(models.ProjectSettings),
		new(models.Catalog),
		new(models.Source),
		new(models.Destination),
		new(models.Job),
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

		// cleanup expired sessions
		if err := conn.Exec(`DELETE FROM session WHERE session_expiry <= NOW()`).Error; err != nil {
			return nil, fmt.Errorf("failed to cleanup expired sessions: %s", err)
		}
	}
	return &Database{conn: conn}, nil
}

// dropLegacyConstraints cleans up constraints left behind by previous ORM versions
// (beego-orm and early GORM) that conflict with the current GORM AutoMigrate:
//
//   - "uni_*" unique constraints (GORM `unique` tag, non-idempotent) → replaced by `uniqueIndex`
//   - "*_fkey" foreign key constraints (beego-orm naming) → GORM recreates with "fk_*" names,
//     causing duplicate constraints that are harmless but add clutter.
//
// All DROP statements use IF EXISTS so this is a no-op on fresh installs.
func dropLegacyConstraints(conn *gorm.DB) error {
	userTable := constants.TableNameMap[constants.UserTable]
	projectSettingsTable := constants.TableNameMap[constants.ProjectSettingsTable]
	sourceTable := constants.TableNameMap[constants.SourceTable]
	destTable := constants.TableNameMap[constants.DestinationTable]
	jobTable := constants.TableNameMap[constants.JobTable]

	constraints := []struct {
		table      string
		constraint string
	}{
		// Legacy GORM unique constraints (non-idempotent `unique` tag → now `uniqueIndex`)
		{projectSettingsTable, fmt.Sprintf("uni_%s_project_id", projectSettingsTable)},
		{userTable, fmt.Sprintf("uni_%s_username", userTable)},
		{userTable, fmt.Sprintf("uni_%s_email", userTable)},

		// Legacy beego-orm unique constraints (PostgreSQL default naming: <table>_<column>_key)
		{projectSettingsTable, fmt.Sprintf("%s_project_id_key", projectSettingsTable)},
		{userTable, fmt.Sprintf("%s_username_key", userTable)},
		{userTable, fmt.Sprintf("%s_email_key", userTable)},

		// GORM FK constraints (fk_<table>_<assoc> naming). FK creation is now disabled
		// via DisableForeignKeyConstraintWhenMigrating because GORM's HasConstraint
		// queries information_schema, which filters by table ownership — if the DB
		// user doesn't own the tables (common with external RDS), GORM can't see
		// existing FK constraints and tries to recreate them, causing "already exists"
		// errors. The app enforces referential integrity in code.
		{sourceTable, fmt.Sprintf("fk_%s_created_by", sourceTable)},
		{sourceTable, fmt.Sprintf("fk_%s_updated_by", sourceTable)},
		{destTable, fmt.Sprintf("fk_%s_created_by", destTable)},
		{destTable, fmt.Sprintf("fk_%s_updated_by", destTable)},
		{jobTable, fmt.Sprintf("fk_%s_source", jobTable)},
		{jobTable, fmt.Sprintf("fk_%s_destination", jobTable)},
		{jobTable, fmt.Sprintf("fk_%s_created_by", jobTable)},
		{jobTable, fmt.Sprintf("fk_%s_updated_by", jobTable)},
	}

	for _, c := range constraints {
		sql := fmt.Sprintf(`ALTER TABLE IF EXISTS %q DROP CONSTRAINT IF EXISTS %q`, c.table, c.constraint)
		if err := conn.Exec(sql).Error; err != nil {
			return fmt.Errorf("failed to drop constraint %s on %s: %w", c.constraint, c.table, err)
		}
	}
	return nil
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

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.OlakePostgresUser, cfg.OlakePostgresPassword),
		Host:   fmt.Sprintf("%s:%s", cfg.OlakePostgresHost, cfg.OlakePostgresPort),
		Path:   "/" + url.PathEscape(cfg.OlakePostgresDBName),
	}

	query := u.Query()
	query.Set("sslmode", cfg.OlakePostgresSSLMode)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
