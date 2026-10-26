// Package sources holds the source connectors the integration test can run against.
// Each source lives in its own folder with its docker-compose.yml.
package sources

import (
	"context"
	"testing"
)

// Source describes how to start a source inside the DinD box and how to seed it.
type Source struct {
	Name       string // shown in logs
	ComposeDir string // folder with docker-compose.yml, relative to server/tests
	Port       string // port the box exposes so the test can reach the source, e.g. "5433/tcp"; empty exposes none
	Namespace  string // <database>_<schema> OLake appends to the job name to name the destination database
	Container  string // the source's container; a workflow stops it to make a sync fail
	ReadyCmd   string // run in the box after compose up, in a subshell; exit 0 once ready, non-zero if it never is; empty skips

	// Seed creates the tables and rows the workflow syncs. host and port reach the source from the
	// test; port is empty when Port is.
	Seed func(ctx context.Context, t *testing.T, host, port string)

	// Exec runs SQL statements on the source in order, for the changes a workflow makes between
	// syncs. host and port are as for Seed.
	Exec func(ctx context.Context, t *testing.T, host, port string, statements ...string)

	// Query runs a query on the source and returns every row as text, NULL as "NULL", so the
	// checks can compare it with the destination. host and port are as for Seed.
	Query func(ctx context.Context, host, port, query string) ([][]string, error)
}
