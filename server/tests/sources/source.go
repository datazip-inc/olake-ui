// Package sources holds the source connectors the integration test can run against.
// Each source lives in its own folder with its docker-compose.yml.
package sources

import (
	"context"
	"testing"
)

// Source describes how to start a source inside the DinD box and how to seed it.
type Source struct {
	Name       string // compose project suffix
	ComposeDir string // folder with docker-compose.yml, relative to server/tests
	Port       string // port the box exposes so the test can reach the source, e.g. "5433/tcp"
	ReadyCmd   string // run in the box after compose up; must exit 0 once the source is ready

	// Seed creates the tables and rows the workflow syncs. Nil when the compose file seeds the source itself. host and port reach the source from the test.
	Seed func(ctx context.Context, t *testing.T, host, port string)
}
