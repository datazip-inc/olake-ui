// Package destinations holds the destinations the integration test can write to.
// Each destination lives in its own folder with its docker-compose.yml.
package destinations

// Destination describes how to start a destination and how Spark reads what it wrote.
//
// The harness checks every destination through Spark Connect on port 15002. So the compose
// file in ComposeDir must start a Spark Connect server published on 15002, with a catalog that
// resolves the names Table returns (the jdbc compose file has the spark-iceberg service for this).
type Destination struct {
	Name       string   // shown in logs
	ComposeDir string   // folder with docker-compose.yml, relative to server/tests
	Services   []string // compose services to start; empty starts all
	ReadyCmd   string   // run after compose up, in a subshell; exit 0 once ready, non-zero if it never is; empty skips
	DataDir    string   // folder, relative to ComposeDir, the compose file bind-mounts for its data; emptied before each run

	// Table returns the name Spark reads a synced table from, given the destination database
	// (or folder) and the source table.
	Table func(database, table string) string
}
