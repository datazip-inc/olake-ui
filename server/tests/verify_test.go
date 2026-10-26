package tests

import (
	"testing"

	"github.com/datazip-inc/olake-ui/server/tests/destinations/jdbc"
	"github.com/datazip-inc/olake-ui/server/tests/sources/postgres"
)

// TestPostgresJDBC creates a Postgres -> Iceberg (JDBC) job through the UI, syncs it, and checks the rows in Iceberg.
func TestPostgresJDBC(t *testing.T) {
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	err := RunWorkflow(t, Workflow{
		Source:      postgres.Source,
		Destination: jdbc.Destination,
		Spec:        "tests/flows/job-end-to-end.spec.ts",
		Job:         "postgres_iceberg_jdbc_job",
		Checks: []Check{
			CountDistinct(postgres.TestTable, 5),
			Count(postgres.TestTable, 5),
		},
	})
	if err != nil {
		t.Errorf("Integration workflow failed: %s", err)
	}
}
