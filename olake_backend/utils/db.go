package utils

import (
	"olake_backend/models"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/lib/pq" 
)

// ConnectDB initializes the database connection
func ConnectDB() {
	// Register the PostgreSQL driver
	orm.RegisterDriver("postgres", orm.DRPostgres)

	// Connection string for PostgreSQL
	connStr := "host=localhost port=5432 user=postgres password=12345678 dbname=olakedb sslmode=disable"
	err := orm.RegisterDataBase("default", "postgres", connStr)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Register models in correct order
	orm.RegisterModel(
		new(models.Source),
		new(models.Destination),
		new(models.Job),
		new(models.User),
	)

	// Create tables if they do not exist
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		panic("Failed to sync database schema: " + err.Error())
	}
}
