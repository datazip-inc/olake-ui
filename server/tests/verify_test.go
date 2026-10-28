package tests

import (
	"slices"
	"strconv"
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

	frAppendTable     = "e2e_fr_append"     // Full Refresh, Append
	frUpsertTable     = "e2e_fr_upsert"     // Full Refresh, Upsert
	typesTable        = "e2e_types"         // Full Refresh + CDC, Append; a column per type, see postgres.Source
	noPKTable         = "e2e_no_pk"         // Full Refresh + CDC, Append; no primary key
	schemaChangeTable = "e2e_schema_change" // Full Refresh + CDC, Append; amount becomes BIGINT, name is dropped
	droppedTable      = "e2e_dropped"       // Full Refresh + CDC, Append; dropped at the source late in the run
)

const postgresSpec = "tests/integration/postgres-iceberg.spec.ts"

// The source columns before and after e2e_new_columns gets extra_col, and e2e_schema_change's
// after it loses name.
var (
	baseColumns    = []string{"id", "name", "amount"}
	extraColumns   = []string{"id", "name", "amount", "extra_col"}
	changedColumns = []string{"id", "amount"}
)

// typeColumns are e2e_types' columns as OLake writes them. JSON, arrays and text arrive as the
// source shows them; a NUMERIC becomes a float, so 0.50 reads 0.5; a DATE becomes a timestamp at
// midnight, and a TIMESTAMPTZ a timestamp in UTC, which Spark shows without an offset.
var typeColumns = []SourceColumn{
	Col("id"),
	{Name: "d", Source: "CAST(CAST(d AS TIMESTAMP) AS TEXT)"},
	Col("ts"),
	{Name: "tstz", Source: "CAST(tstz AT TIME ZONE 'UTC' AS TEXT)"},
	{Name: "num", Source: "CAST(CAST(num AS DOUBLE PRECISION) AS TEXT)"},
	Col("js"),
	Col("arr"),
	Col("txt"),
}

// cdcTypeColumns are typeColumns but arr, for rows CDC wrote. OLake writes a Postgres array that
// CDC reads as a JSON list of its text, {-1,0,1} as ["{-1,0,1}"], unlike the full load, which keeps
// {-1,0,1}. Compare arr through CDC too once OLake writes both the same.
var cdcTypeColumns = slices.DeleteFunc(slices.Clone(typeColumns), func(c SourceColumn) bool {
	return c.Name == "arr"
})

// filterWhere is e2e_filter's Data Filter, as the source query that gives the same rows.
const filterWhere = "amount > 20 AND id < 5"

// TestPostgres runs the Postgres → Iceberg (JDBC) workflow: one job with a stream per feature,
// created and driven through the UI, and its data checked in Iceberg through Spark after every
// sync. See postgres-iceberg.spec.ts for the UI side of each step.
func TestPostgres(t *testing.T) {
	r := StartWorkflow(t, &Workflow{
		Source:      postgres.Source,
		Destination: jdbc.Destination,
		Job:         "e2e_postgres_jdbc",
	})
	// step runs the spec's test tagged tag. runStep is step for a test that starts a run of the
	// job: it passes E2E_RUNS, how many runs the job's history lists once the run has started.
	runs := 0
	step := func(tag string) { r.Playwright(postgresSpec, tag, nil) }
	runStep := func(tag string) {
		runs++
		r.Playwright(postgresSpec, tag, map[string]string{"E2E_RUNS": strconv.Itoa(runs)})
	}

	step("@create-job")

	// Sync 1: the full loads. CDC Only starts at the current WAL position, so it writes nothing.
	runStep("@first-sync")
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

		// Row by row against the source.
		r.MatchesSource(cdcTable, baseColumns, ""),
		r.MatchesSource(upsertEqTable, baseColumns, ""),
		r.MatchesSource(upsertDVTable, baseColumns, ""),
		r.MatchesSource(filterTable, baseColumns, filterWhere),
		r.MatchesSource(newColumnsTable, baseColumns, ""),

		Ops(frAppendTable, map[string]int64{"r": 5}),
		r.MatchesSource(frAppendTable, baseColumns, ""),
		Ops(frUpsertTable, map[string]int64{"r": 5}),
		OneRowPerID(frUpsertTable),
		r.MatchesSource(frUpsertTable, baseColumns, ""),
		r.MatchesSourceAs(typesTable, typeColumns, ""),
		Ops(noPKTable, map[string]int64{"r": 5}),
		r.MatchesSource(noPKTable, baseColumns, ""),
		Ops(schemaChangeTable, map[string]int64{"r": 5}),
		r.MatchesSource(schemaChangeTable, baseColumns, ""),
		Ops(droppedTable, map[string]int64{"r": 5}),
		r.MatchesSource(droppedTable, baseColumns, ""),
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

		// Full Refresh reads these again in sync 2, after the driver drops their tables.
		"INSERT INTO e2e_fr_append VALUES (6, 'name_6', 60)",
		"UPDATE e2e_fr_append SET name = 'updated_1' WHERE id = 1",
		"DELETE FROM e2e_fr_append WHERE id = 2",
		"INSERT INTO e2e_fr_upsert VALUES (6, 'name_6', 60)",
		"UPDATE e2e_fr_upsert SET name = 'updated_1' WHERE id = 1",
		"DELETE FROM e2e_fr_upsert WHERE id = 2",

		// Each type again, through CDC this time.
		`INSERT INTO e2e_types VALUES (4, '2024-02-29', '2024-02-29 23:59:59.123456', '2024-02-29 12:00:00-08:00',
			0.50, '{"nested": {"a": [1, 2]}}', '{-1,0,1}', 'cdc ✓ ünïcödé')`,
		"UPDATE e2e_types SET num = 99.99, txt = 'updated ✓' WHERE id = 3",

		"INSERT INTO e2e_no_pk VALUES (6, 'name_6', 60)",
		"UPDATE e2e_no_pk SET name = 'updated_1' WHERE id = 1",
		"DELETE FROM e2e_no_pk WHERE id = 2",

		// INT to BIGINT is a type change Iceberg allows; the new values do not fit in an INT.
		"ALTER TABLE e2e_schema_change ALTER COLUMN amount TYPE BIGINT",
		"ALTER TABLE e2e_schema_change DROP COLUMN name",
		"INSERT INTO e2e_schema_change VALUES (6, 6000000000)",
		"UPDATE e2e_schema_change SET amount = 7000000000 WHERE id = 1",
	)

	// Sync 2: insert, update and delete arrive through CDC.
	runStep("@sync")
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

		// Row by row against the source. CDC Only holds only the rows inserted after sync 1.
		r.MatchesSource(cdcTable, baseColumns, ""),
		r.MatchesSource(strictCDCTable, baseColumns, "id >= 6"),
		r.MatchesSource(upsertEqTable, baseColumns, ""),
		r.MatchesSource(upsertDVTable, baseColumns, ""),
		r.MatchesSource(filterTable, baseColumns, filterWhere),
		r.MatchesSource(newColumnsTable, extraColumns, ""),

		// Full Refresh rewrote both tables: the source as it is now, no duplicates.
		Count(frAppendTable, 5),
		Ops(frAppendTable, map[string]int64{"r": 5}),
		r.MatchesSource(frAppendTable, baseColumns, ""),
		Count(frUpsertTable, 5),
		Ops(frUpsertTable, map[string]int64{"r": 5}),
		OneRowPerID(frUpsertTable),
		r.MatchesSource(frUpsertTable, baseColumns, ""),

		r.MatchesSourceAs(typesTable, cdcTypeColumns, ""),

		Ops(noPKTable, map[string]int64{"r": 5, "c": 1, "u": 1, "d": 1}),
		r.MatchesSource(noPKTable, baseColumns, ""),

		Value(schemaChangeTable, 1, "amount", "7000000000"),
		Value(schemaChangeTable, 6, "amount", "6000000000"),
		r.MatchesSource(schemaChangeTable, changedColumns, ""),

		Ops(droppedTable, map[string]int64{"r": 5}),
		r.MatchesSource(droppedTable, baseColumns, ""),
	)

	// Clear Destination drops every table of the job.
	runStep("@clear-destination")
	r.Check("after clear destination",
		Absent(cdcTable),
		Absent(strictCDCTable),
		Absent(upsertEqTable),
		Absent(upsertDVTable),
		Absent(filterTable),
		Absent(newColumnsTable),
		Absent(frAppendTable),
		Absent(frUpsertTable),
		Absent(typesTable),
		Absent(noPKTable),
		Absent(schemaChangeTable),
		Absent(droppedTable),
	)

	// The next sync rebuilds the tables in full from the source as it is now.
	runStep("@sync")
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

		// Row by row against the source.
		r.MatchesSource(cdcTable, baseColumns, ""),
		r.MatchesSource(upsertEqTable, baseColumns, ""),
		r.MatchesSource(upsertDVTable, baseColumns, ""),
		r.MatchesSource(filterTable, baseColumns, filterWhere),
		r.MatchesSource(newColumnsTable, extraColumns, ""),

		Ops(frAppendTable, map[string]int64{"r": 5}),
		r.MatchesSource(frAppendTable, baseColumns, ""),
		Ops(frUpsertTable, map[string]int64{"r": 5}),
		r.MatchesSource(frUpsertTable, baseColumns, ""),
		r.MatchesSourceAs(typesTable, typeColumns, ""),
		Ops(noPKTable, map[string]int64{"r": 5}),
		r.MatchesSource(noPKTable, baseColumns, ""),
		r.MatchesSource(schemaChangeTable, changedColumns, ""),
		Ops(droppedTable, map[string]int64{"r": 5}),
		r.MatchesSource(droppedTable, baseColumns, ""),
	)

	// Edit Streams: Target Query Engines decide the upsert types again; the step leaves unsaved.
	step("@edit-query-engines")

	// Edit Streams: e2e_upsert_dv goes from Deletion Vector (Iceberg v3) to Positional Deletes (v2),
	// so its table, and only its, is cleared.
	runStep("@edit-upsert-type")
	r.Check("after the upsert type change",
		Absent(upsertDVTable),
		LiveIDs(upsertEqTable, 1, 3, 4, 5),
		LiveIDs(cdcTable, 1, 3, 4, 5, 6),
	)

	// The next sync loads it again in full, now Iceberg v2; then CDC changes it with position deletes.
	runStep("@sync")
	r.Check("after the re-sync",
		Ops(upsertDVTable, map[string]int64{"r": 4}),
		LiveIDs(upsertDVTable, 1, 3, 4, 5),
		OneRowPerID(upsertDVTable),
		FormatVersion(upsertDVTable, 2),
		r.MatchesSource(upsertDVTable, baseColumns, ""),
	)
	r.SourceSQL(
		"UPDATE e2e_upsert_dv SET name = 'updated_3' WHERE id = 3",
		"DELETE FROM e2e_upsert_dv WHERE id = 4",
	)
	runStep("@sync")
	r.Check("after position deletes",
		LiveIDs(upsertDVTable, 1, 3, 5),
		OneRowPerID(upsertDVTable),
		Value(upsertDVTable, 3, "name", "updated_3"),
		DeleteFiles(upsertDVTable, PositionDeletes),
		FormatVersion(upsertDVTable, 2),
		r.MatchesSource(upsertDVTable, baseColumns, ""),
	)

	// A selected table dropped at the source: the publication no longer covers it, so CDC syncs fail
	// until Edit Streams, which no longer discovers it, saves the job without it.
	r.SourceSQL("DROP TABLE e2e_dropped")
	runStep("@dropped-table")
	runStep("@sync")
	r.Check("after the dropped table's stream is gone",
		// Its destination table stays as it was.
		Ops(droppedTable, map[string]int64{"r": 5}),
		LiveIDs(droppedTable, 1, 2, 3, 4, 5),
	)

	// UI-only steps.
	step("@stream-rules")

	// Alerts: a failed sync posts "Sync Failure Detected!" to the saved webhook URL.
	url := r.StartWebhookReceiver()
	r.Playwright(postgresSpec, "@alerts", map[string]string{"E2E_WEBHOOK_URL": url})

	// Changes made before the source goes down: the failed sync cannot read them, so the sync after
	// it must, exactly once.
	r.SourceSQL(
		"INSERT INTO e2e_cdc VALUES (7, 'name_7', 70)",
		"UPDATE e2e_cdc SET name = 'updated_3' WHERE id = 3",
	)
	r.StopSource()
	runStep("@failed-sync")
	step("@discovery-failure")
	r.WaitForWebhook("Sync Failure Detected!", 5*time.Minute)

	// Recovery: with the source back, discovery works again, and CDC resumes where it stopped.
	r.StartSource()
	runStep("@recovery")
	r.Check("after recovery",
		Ops(cdcTable, map[string]int64{"r": 5, "c": 1, "u": 1}),
		LiveIDs(cdcTable, 1, 3, 4, 5, 6, 7),
		Value(cdcTable, 3, "name", "updated_3"),
		r.MatchesSource(cdcTable, baseColumns, ""),
	)

	// The history of every run, over two pages, and the newest run's logs.
	r.Playwright(postgresSpec, "@history", map[string]string{"E2E_RUNS": strconv.Itoa(runs)})
}
