package tests

import (
	"os"
	"testing"

	"github.com/datazip-inc/olake-ui/server/tests/destinations/jdbc"
	"github.com/datazip-inc/olake-ui/server/tests/destinations/s3"
	"github.com/datazip-inc/olake-ui/server/tests/sources/mongodb"
	"github.com/datazip-inc/olake-ui/server/tests/sources/postgres"
)

// TestPostgres drives the Postgres -> Iceberg (JDBC) journey through the UI: six feature tables in one
// job, synced twice with source changes in between, then cleared and synced again. The spec runs in
// phases so the tables can be looked at before the clear, after it and after the rebuild.
func TestPostgres(t *testing.T) {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	// the spec sets the job's destination database prefix to e2e_features; dynamic names add the schema
	const database = "e2e_features_public"
	var allTables = []string{
		postgres.TableCDC, postgres.TableStrictCDC, postgres.TableUpsertEq,
		postgres.TableUpsertDV, postgres.TableFilter, postgres.TableNewColumns,
	}
	dropped := make([]Check, 0, len(allTables))
	for _, table := range allTables {
		dropped = append(dropped, Dropped(table))
	}

	err := DinDTestContainer(t, Workflow{
		Source:      postgres.Source,
		Destination: jdbc.Destination,
		Spec:        "tests/flows/postgres-jdbc.spec.ts",
		Database:    database,
		Phases: []Phase{
			{
				// everything up to and including the second sync
				Grep: "create a CDC source|stream rules|create the features job|first sync|task logs|second sync",
				Checks: []Check{
					// insert, update and delete after the first sync arrive as r=5 c=1 u=1 d=1
					OpTypes(postgres.TableCDC, map[string]int{"r": 5, "c": 1, "u": 1, "d": 1}),
					// the first sync wrote nothing, the second only the 2 new rows
					Rows(postgres.TableStrictCDC, "1:10", "2:20"),
					Rows(postgres.TableUpsertEq, "1:99", "3:30", "4:40", "5:50"),
					Rows(postgres.TableUpsertDV, "1:99", "3:30", "4:40", "5:50"),
					DeletionVector(postgres.TableUpsertDV),
					// id 2 started matching amount > 20 AND id < 5, ids 6 and 7 stay out
					Rows(postgres.TableFilter, "2:25", "3:30"),
					Rows(postgres.TableNewColumns, "1:10", "2:20", "3:30"),
					NewColumn(postgres.TableNewColumns, "extra"),
				},
			},
			{
				// Clear Destination on the job drops every table it wrote
				Grep:   "clear destination",
				Checks: dropped,
			},
			{
				// the next sync rebuilds every table in full from the current source
				Grep: "rebuilds every table",
				Checks: []Check{
					OpTypes(postgres.TableCDC, map[string]int{"r": 5}),
					// CDC Only has no snapshot, so after the clear it starts again with nothing
					Rows(postgres.TableStrictCDC),
					Rows(postgres.TableUpsertEq, "1:99", "3:30", "4:40", "5:50"),
					Rows(postgres.TableUpsertDV, "1:99", "3:30", "4:40", "5:50"),
					Rows(postgres.TableFilter, "2:25", "3:30"),
					Rows(postgres.TableNewColumns, "1:10", "2:20", "3:30"),
					NewColumn(postgres.TableNewColumns, "extra"),
				},
			},
			{
				// no data to check: the webhook and the failed job are looked at in the UI
				Grep: "webhook alerts|failed sync",
			},
		},
	})
	if err != nil {
		t.Errorf("Error in Docker in Docker container start up: %s", err)
	}
}

// TestMongoDB syncs MongoDB collections to S3 parquet, one per behaviour: normalization on
// (fields as columns) and off (one data column), an incremental cursor with a fallback cursor, a
// partition regex, and a collection added later through Edit Streams.
func TestMongoDB(t *testing.T) {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	const (
		names      = `SELECT DISTINCT name FROM %s ORDER BY name`
		normalized = `SELECT DISTINCT name, age FROM %s ORDER BY name`
		raw        = `SELECT DISTINCT get_json_object(data, '$.name'), get_json_object(data, '$.age') FROM %s ORDER BY 1`
		partitions = `SELECT DISTINCT regexp_extract(input_file_name(), 'category=([^/]+)/', 1) FROM %s ORDER BY 1`
	)
	err := DinDTestContainer(t, Workflow{
		Source:      mongodb.Source,
		Destination: s3.Destination,
		Spec:        "tests/flows/mongodb-parquet.spec.ts",
		// the spec gives the job a custom S3 folder, so every table is written directly under it
		Database: "e2e_mongo_folder",
		Phases: []Phase{{Checks: []Check{
			// every full refresh appends, so the same rows can be there twice
			DataColumns(mongodb.CollectionFullRefresh, "_id", "address", "age", "name", "tags"),
			Query(mongodb.CollectionFullRefresh, normalized, "alice:30", "bob:25", "carol:41"),
			DataColumns(mongodb.CollectionNormOff, "data"),
			Query(mongodb.CollectionNormOff, raw, "alice:30", "bob:25", "carol:41"),
			// i4 came in through the cursor, i5 (no updated_at) through the fallback cursor
			Query(mongodb.CollectionIncremental, names, "i1", "i2", "i3", "i4", "i5"),
			Query(mongodb.CollectionPartition, partitions, "a", "b"),
			Query(mongodb.CollectionAdded, names, "x1", "x2"),
		}}},
	})
	if err != nil {
		t.Errorf("Error in Docker in Docker container start up: %s", err)
	}
}
