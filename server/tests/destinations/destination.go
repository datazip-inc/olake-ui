// Package destinations holds the destinations the integration test can write to.
// Each destination lives in its own folder with its docker-compose.yml.
package destinations

// Destination describes how to start a destination inside the DinD box and how Spark reads what it wrote.
type Destination struct {
	Name       string   // compose project suffix
	ComposeDir string   // folder with docker-compose.yml, relative to server/tests
	Services   []string // compose services to start; empty starts all

	// Table returns the name Spark reads a synced table from, given the destination database
	// (or folder) and the source table.
	Table func(database, table string) string

	// SparkInit holds statements run on the Spark session before the checks, for example the
	// s3a settings that parquet files need.
	SparkInit []string
}
