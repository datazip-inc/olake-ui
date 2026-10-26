package tests

import (
	"testing"

	"github.com/datazip-inc/olake-ui/server/tests/destinations/jdbc"
	"github.com/datazip-inc/olake-ui/server/tests/sources/postgres"
)

// TestPostgresJDBC creates a Postgres -> Iceberg (JDBC) job through the UI, syncs it, and checks the rows in Iceberg.
func TestPostgresJDBC(t *testing.T) {
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	r := StartWorkflow(t, Workflow{
		Source:      postgres.Source,
		Destination: jdbc.Destination,
		Job:         "postgres_iceberg_jdbc_job",
	})
	r.Playwright("tests/flows/job-end-to-end.spec.ts", "", nil)
	r.Check("after sync",
		CountDistinct(postgres.TestTable, 5),
		Count(postgres.TestTable, 5),
	)
}
