// Package jdbc is the Apache Iceberg destination with a JDBC catalog, data in MinIO.
package jdbc

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
)

// Destination starts MinIO, the catalog Postgres and Spark; Spark reads the tables through olake_iceberg.
var Destination = destinations.Destination{
	Name:       "jdbc",
	ComposeDir: "destinations/jdbc",
	Services:   []string{"minio", "mc", "postgres", "spark-iceberg"},
	Table: func(database, table string) string {
		return fmt.Sprintf("olake_iceberg.%s.%s", database, table)
	},
}
