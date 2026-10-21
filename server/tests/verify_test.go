package tests

import (
	"os"
	"testing"

	"github.com/datazip-inc/olake-ui/server/tests/destinations/jdbc"
	"github.com/datazip-inc/olake-ui/server/tests/sources/postgres"
)

// TestPostgres creates a Postgres -> Iceberg (JDBC) job through the UI, syncs it, and checks the rows in Iceberg.
func TestPostgres(t *testing.T) {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	err := DinDTestContainer(t, Workflow{
		Source:      postgres.Source,
		Destination: jdbc.Destination,
		Spec:        "tests/flows/job-end-to-end.spec.ts",
		Database:    "postgres_iceberg_jdbc_job_postgres_public",
		Checks:      []Check{CountDistinct(postgres.TestTable, 5)},
	})
	if err != nil {
		t.Errorf("Error in Docker in Docker container start up: %s", err)
	}
}
