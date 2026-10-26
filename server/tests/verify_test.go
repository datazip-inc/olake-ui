package tests

import (
	"testing"
	"time"

	"github.com/datazip-inc/olake-ui/server/tests/destinations/jdbc"
	"github.com/datazip-inc/olake-ui/server/tests/sources/postgres"
)

// The e2e tables of the Postgres workflow; postgres.Source seeds each with rows 1-5:
// id, name 'name_<id>' and amount <id>*10.
const (
	cdcTable        = "e2e_cdc"         // Full Refresh + CDC, Append
	strictCDCTable  = "e2e_strict_cdc"  // CDC Only, Append
	upsertEqTable   = "e2e_upsert_eq"   // Full Refresh + CDC, Upsert, Equality Deletes
	upsertDVTable   = "e2e_upsert_dv"   // Full Refresh + CDC, Upsert, Deletion Vector
	filterTable     = "e2e_filter"      // Full Refresh + CDC, Append, Data Filter amount > 20 AND id < 5
	newColumnsTable = "e2e_new_columns" // Full Refresh + CDC, Append, Sync new columns on
)

const postgresSpec = "tests/integration/postgres-iceberg.spec.ts"

// TestPostgres runs the Postgres → Iceberg (JDBC) workflow: one job with a stream per feature,
// created and driven through the UI, and its data checked in Iceberg through Spark after every
// sync. See postgres-iceberg.spec.ts for the UI side of each step.
func TestPostgres(t *testing.T) {
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	r := StartWorkflow(t, Workflow{
		Source:      postgres.Source,
		Destination: jdbc.Destination,
		Job:         "e2e_postgres_jdbc",
	})

	r.Playwright(postgresSpec, "@create-job", nil)

	// Sync 1: the full loads. CDC Only starts at the current WAL position, so it writes nothing.
	r.Playwright(postgresSpec, "@first-sync", nil)
	r.Check("after sync 1",
		Ops(cdcTable, map[string]int64{"r": 5}),
		CountOrAbsent(strictCDCTable, 0),
		LiveIDs(upsertEqTable, 1, 2, 3, 4, 5),
		OneRowPerID(upsertEqTable),
		LiveIDs(upsertDVTable, 1, 2, 3, 4, 5),
		OneRowPerID(upsertDVTable),
		FormatVersion(upsertDVTable, 3),
		LiveIDs(filterTable, 3, 4),
		LiveIDs(newColumnsTable, 1, 2, 3, 4, 5),
		Columns(newColumnsTable, []string{"id", "name", "amount"}, []string{"extra_col"}),
	)

	// Source changes, read by CDC in sync 2.
	r.SourceSQL(
		"INSERT INTO e2e_cdc VALUES (6, 'name_6', 60)",
		"UPDATE e2e_cdc SET name = 'updated_1' WHERE id = 1",
		"DELETE FROM e2e_cdc WHERE id = 2",

		"INSERT INTO e2e_strict_cdc VALUES (6, 'name_6', 60), (7, 'name_7', 70)",

		"UPDATE e2e_upsert_eq SET name = 'updated_1' WHERE id = 1",
		"DELETE FROM e2e_upsert_eq WHERE id = 2",

		"UPDATE e2e_upsert_dv SET name = 'updated_1' WHERE id = 1",
		"DELETE FROM e2e_upsert_dv WHERE id = 2",

		// Row 6 is outside the filter (id < 5 fails); row 2 starts matching (amount 25 > 20).
		"INSERT INTO e2e_filter VALUES (6, 'name_6', 60)",
		"UPDATE e2e_filter SET amount = 25 WHERE id = 2",

		"ALTER TABLE e2e_new_columns ADD COLUMN extra_col INT",
		"INSERT INTO e2e_new_columns VALUES (6, 'name_6', 60, 99)",
	)

	// Sync 2: insert, update and delete arrive through CDC.
	r.Playwright(postgresSpec, "@second-sync", nil)
	r.Check("after sync 2",
		Ops(cdcTable, map[string]int64{"r": 5, "c": 1, "u": 1, "d": 1}),
		Value(cdcTable, 1, "name", "updated_1"),

		LiveIDs(strictCDCTable, 6, 7),
		Ops(strictCDCTable, map[string]int64{"c": 2}),

		LiveIDs(upsertEqTable, 1, 3, 4, 5),
		OneRowPerID(upsertEqTable),
		Value(upsertEqTable, 1, "name", "updated_1"),
		DeleteFiles(upsertEqTable, EqualityDeletes),

		LiveIDs(upsertDVTable, 1, 3, 4, 5),
		OneRowPerID(upsertDVTable),
		Value(upsertDVTable, 1, "name", "updated_1"),
		FormatVersion(upsertDVTable, 3),
		DeleteFiles(upsertDVTable, DeletionVectors),

		LiveIDs(filterTable, 2, 3, 4),
		Value(filterTable, 2, "amount", "25"),

		Columns(newColumnsTable, []string{"id", "name", "amount", "extra_col"}, nil),
		LiveIDs(newColumnsTable, 1, 2, 3, 4, 5, 6),
		Value(newColumnsTable, 6, "extra_col", "99"),
	)

	// Clear Destination drops every table of the job.
	r.Playwright(postgresSpec, "@clear-destination", nil)
	r.Check("after clear destination",
		Absent(cdcTable),
		Absent(strictCDCTable),
		Absent(upsertEqTable),
		Absent(upsertDVTable),
		Absent(filterTable),
		Absent(newColumnsTable),
	)

	// The next sync rebuilds the tables in full from the source as it is now.
	r.Playwright(postgresSpec, "@rebuild-sync", nil)
	r.Check("after rebuild",
		Count(cdcTable, 5),
		Ops(cdcTable, map[string]int64{"r": 5}),
		LiveIDs(cdcTable, 1, 3, 4, 5, 6),
		Value(cdcTable, 1, "name", "updated_1"),

		CountOrAbsent(strictCDCTable, 0),

		LiveIDs(upsertEqTable, 1, 3, 4, 5),
		OneRowPerID(upsertEqTable),
		Value(upsertEqTable, 1, "name", "updated_1"),

		LiveIDs(upsertDVTable, 1, 3, 4, 5),
		OneRowPerID(upsertDVTable),
		FormatVersion(upsertDVTable, 3),

		LiveIDs(filterTable, 2, 3, 4),

		Columns(newColumnsTable, []string{"id", "name", "amount", "extra_col"}, nil),
		LiveIDs(newColumnsTable, 1, 2, 3, 4, 5, 6),
		Value(newColumnsTable, 6, "extra_col", "99"),
	)

	// UI-only steps.
	r.Playwright(postgresSpec, "@history", nil)
	r.Playwright(postgresSpec, "@stream-rules", nil)

	// Alerts: a failed sync posts "Sync Failure Detected!" to the saved webhook URL.
	url := r.StartWebhookReceiver()
	r.Playwright(postgresSpec, "@alerts", map[string]string{"E2E_WEBHOOK_URL": url})
	r.StopSource()
	r.Playwright(postgresSpec, "@failed-sync", nil)
	r.WaitForWebhook("Sync Failure Detected!", 5*time.Minute)
}
