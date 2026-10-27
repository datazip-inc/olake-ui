package tests

import (
	"strconv"
	"testing"

	"github.com/datazip-inc/olake-ui/server/tests/destinations/s3"
	"github.com/datazip-inc/olake-ui/server/tests/sources/mongodb"
)

// The collections of the MongoDB workflow, which init.js seeds (server/tests/sources/mongodb).
const (
	fullRefreshColl = "e2e_full_refresh" // Full Refresh, Normalization on; deselected late in the run
	normOffColl     = "e2e_norm_off"     // Full Refresh, Normalization off; turned on in Edit Streams
	incrementalColl = "e2e_incremental"  // Incremental on updated_at, fallback cursor created_at
	partitionColl   = "e2e_partition"    // Full Refresh, partition regex /category={category,default_category,}
	columnsColl     = "e2e_columns"      // Full Refresh; secret deselected at creation, note in Edit Streams
	filterOrColl    = "e2e_filter_or"    // Incremental, Data Filter category = "a" or amount <= 10
	filterAndColl   = "e2e_filter_and"   // Incremental, Data Filter amount != 30 and amount >= 20
	partitionTsColl = "e2e_partition_ts" // Full Refresh, partition regex /year={event_at,,YYYY}/month={event_at,,MM}
	addedColl       = "e2e_added"        // added to the job in Edit Streams
)

const mongoSpec = "tests/integration/mongodb-parquet.spec.ts"

// mongoFolder is the custom S3 folder the spec gives the job: every collection is written to
// s3://warehouse/e2e/<mongoFolder>/<collection>/.
const mongoFolder = "e2e_mongo_folder"

// Queries for Rows, %s being the collection's parquet files.
const (
	namesQuery  = "SELECT name FROM %s ORDER BY name"
	peopleQuery = "SELECT name, age FROM %s ORDER BY name"
	// With Normalization off the document is the JSON in the data column.
	rawPeopleQuery = "SELECT get_json_object(data, '$.name'), get_json_object(data, '$.age') FROM %s ORDER BY 1"
	// The folder a partition regex put each document in, relative to the collection's folder.
	categoryFoldersQuery = "SELECT DISTINCT regexp_extract(input_file_name(), '/(category=[^/]+)/', 1), category FROM %s ORDER BY 1"
	monthFoldersQuery    = "SELECT regexp_extract(input_file_name(), '/(year=[^/]+/month=[^/]+)/', 1), name FROM %s ORDER BY name"
)

// TestMongoDB runs the MongoDB → S3 (parquet) workflow: one job with a stream per feature,
// created and driven through the UI, and its parquet files checked through Spark after every
// sync. See mongodb-parquet.spec.ts for the UI side of each step.
func TestMongoDB(t *testing.T) {
	r := StartWorkflow(t, &Workflow{
		Source:      mongodb.Source,
		Destination: s3.Destination,
		Job:         "e2e_mongo_parquet",
		Database:    mongoFolder,
	})
	// step runs the spec's test tagged tag. runs, for a step that starts a run of the job, is how
	// many runs the job's history lists once it has; @schedule gets the runs before it.
	step := func(tag string, runs int) {
		var env map[string]string
		if runs > 0 {
			env = map[string]string{"E2E_RUNS": strconv.Itoa(runs)}
		}
		r.Playwright(mongoSpec, tag, env)
	}

	step("@create-job", 0)
	step("@stream-rules", 0)

	// Sync 1: every collection in full; the filters and the partition regexes apply already.
	step("@first-sync", 1)
	r.Check("after sync 1",
		Columns(fullRefreshColl, []string{"name", "age", "address", "tags"}, []string{"data"}),
		Rows(fullRefreshColl, "people", peopleQuery, "alice:30", "bob:25", "carol:41"),
		Columns(normOffColl, []string{"data"}, []string{"name", "age"}),
		Rows(normOffColl, "people in data", rawPeopleQuery, "alice:30", "bob:25", "carol:41"),
		Rows(incrementalColl, "names", namesQuery, "i1", "i2", "i3"),
		Rows(partitionColl, "folders", categoryFoldersQuery, "category=a:a", "category=b:b"),
		Columns(columnsColl, []string{"name", "amount", "note"}, []string{"secret"}),
		Rows(filterOrColl, "names", namesQuery, "f1", "f2", "f4"),
		Rows(filterAndColl, "names", namesQuery, "g1", "g3"),
		Rows(partitionTsColl, "folders", monthFoldersQuery,
			"year=2024/month=01:t1", "year=2024/month=01:t2", "year=2025/month=06:t3"),
		Absent(addedColl),
	)

	// Source changes, read by sync 2.
	r.ChangeSource(
		// i4 comes in through the cursor, i5, with no updated_at, through the fallback cursor.
		`db.e2e_incremental.insertMany([
			{ name: "i4", updated_at: new Date("2025-01-01T00:00:00Z"), created_at: new Date("2025-01-01T00:00:00Z") },
			{ name: "i5", updated_at: null, created_at: new Date("2025-01-02T00:00:00Z") },
		])`,
		// The filters apply to new documents too: f5 and g6 match, f6 and g5 do not.
		`db.e2e_filter_or.insertMany([
			{ name: "f5", category: "a", amount: 99, updated_at: new Date("2025-01-01T00:00:00Z") },
			{ name: "f6", category: "b", amount: 11, updated_at: new Date("2025-01-01T00:00:00Z") },
		])`,
		`db.e2e_filter_and.insertMany([
			{ name: "g5", amount: 30, updated_at: new Date("2025-01-01T00:00:00Z") },
			{ name: "g6", amount: 25, updated_at: new Date("2025-01-01T00:00:00Z") },
		])`,
		// Full Refresh reads every collection again: dave arrives, nothing twice.
		`db.e2e_full_refresh.insertOne({ name: "dave", age: 52, address: { city: "oslo" }, tags: [] })`,
		`db.e2e_columns.insertOne({ name: "c3", amount: 3, secret: "s3", note: "n3" })`,
		`db.e2e_partition_ts.insertOne({ name: "t4", event_at: new Date("2025-06-15T10:00:00Z") })`,
	)

	// Edit Streams: a new frequency, re-discovery, e2e_added in, note out of e2e_columns. Adding
	// e2e_added clears its folder: run 2.
	step("@edit-streams", 2)
	step("@sync", 3)
	r.Check("after sync 2",
		Rows(incrementalColl, "names", namesQuery, "i1", "i2", "i3", "i4", "i5"),
		Rows(filterOrColl, "names", namesQuery, "f1", "f2", "f4", "f5"),
		Rows(filterAndColl, "names", namesQuery, "g1", "g3", "g6"),
		Rows(fullRefreshColl, "people", peopleQuery, "alice:30", "bob:25", "carol:41", "dave:52"),
		// Full Refresh wrote e2e_columns again, so neither deselected column is in any file.
		Columns(columnsColl, []string{"name", "amount"}, []string{"secret", "note"}),
		Rows(columnsColl, "names", namesQuery, "c1", "c2", "c3"),
		Columns(addedColl, []string{"name", "age"}, nil),
		Rows(addedColl, "names", namesQuery, "x1", "x2"),
		Rows(partitionTsColl, "folders", monthFoldersQuery,
			"year=2024/month=01:t1", "year=2024/month=01:t2", "year=2025/month=06:t3", "year=2025/month=06:t4"),
	)

	// Edit Streams: Normalization on for e2e_norm_off clears its folder, and only its: run 4.
	step("@edit-stream-change", 4)
	r.Check("after the stream change",
		Absent(normOffColl),
		Rows(fullRefreshColl, "people", peopleQuery, "alice:30", "bob:25", "carol:41", "dave:52"),
		Rows(incrementalColl, "names", namesQuery, "i1", "i2", "i3", "i4", "i5"),
	)
	step("@sync", 5)
	r.Check("after the re-sync",
		Columns(normOffColl, []string{"name", "age"}, []string{"data"}),
		Rows(normOffColl, "people", peopleQuery, "alice:30", "bob:25", "carol:41"),
	)

	// Edit Streams: e2e_full_refresh leaves the job, so the next sync neither reads nor drops it.
	step("@deselect-stream", 0)
	r.ChangeSource(`db.e2e_full_refresh.insertOne({ name: "erin", age: 33 })`)
	step("@sync", 6)
	r.Check("after the deselect",
		Rows(fullRefreshColl, "people", peopleQuery, "alice:30", "bob:25", "carol:41", "dave:52"),
	)

	// Every minute: the schedule's runs 7 and 8, with the job paused in between.
	step("@schedule", 6)
}
