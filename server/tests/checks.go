package tests

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/spark-connect-go/v35/spark/sql"
)

// Check verifies one property of a destination table through Spark. Run returns an error rather
// than failing the test, so the harness can retry it until a sync's commit is visible.
type Check struct {
	Name string
	Run  func(ctx context.Context, spark sql.SparkSession, table func(name string) string) error
}

// checkTimeout and checkInterval bound how long a check is retried before it fails.
const (
	checkTimeout  = 2 * time.Minute
	checkInterval = 5 * time.Second
)

// identPattern limits the table and column names the checks put into SQL.
var identPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func ident(name string) string {
	if !identPattern.MatchString(name) {
		panic(fmt.Sprintf("invalid identifier %q", name))
	}
	return name
}

// runChecks runs every check as a subtest of t, retrying each until it passes or checkTimeout
// runs out. It fails t once all checks have run if any of them failed.
func runChecks(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string, checks []Check) {
	t.Helper()
	failed := false
	for _, check := range checks {
		ok := t.Run(check.Name, func(t *testing.T) {
			deadline := time.Now().Add(checkTimeout)
			for {
				err := check.Run(ctx, spark, table)
				if err == nil {
					t.Logf("✅ %s", check.Name)
					return
				}
				if time.Now().After(deadline) || ctx.Err() != nil {
					t.Fatalf("❌ %s: %v", check.Name, err)
				}
				time.Sleep(checkInterval)
			}
		})
		failed = failed || !ok
	}
	if failed {
		t.FailNow()
	}
}

// waitForSpark runs SELECT 1 until Spark Connect answers, for up to checkTimeout.
func waitForSpark(ctx context.Context, spark sql.SparkSession) error {
	deadline := time.Now().Add(checkTimeout)
	for {
		_, err := query(ctx, spark, "SELECT 1")
		if err == nil || time.Now().After(deadline) || ctx.Err() != nil {
			return err
		}
		time.Sleep(checkInterval)
	}
}

// query runs q and returns every row as strings; NULL becomes "NULL".
func query(ctx context.Context, spark sql.SparkSession, q string) ([][]string, error) {
	df, err := spark.Sql(ctx, q)
	if err != nil {
		return nil, err
	}
	rows, err := df.Collect(ctx)
	if err != nil {
		return nil, err
	}
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		values := make([]string, 0, row.Len())
		for _, v := range row.Values() {
			if v == nil {
				values = append(values, "NULL")
			} else {
				values = append(values, fmt.Sprint(v))
			}
		}
		out = append(out, values)
	}
	return out, nil
}

// isTableMissing reports whether err says the table, or its database, does not exist.
func isTableMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, marker := range []string{"TABLE_OR_VIEW_NOT_FOUND", "NoSuchTableException", "SCHEMA_NOT_FOUND", "NoSuchNamespaceException"} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// refreshed queries tbl after REFRESH TABLE, so Spark does not answer from a cached snapshot.
func refreshed(ctx context.Context, spark sql.SparkSession, tbl, q string) ([][]string, error) {
	if _, err := spark.Sql(ctx, "REFRESH TABLE "+tbl); err != nil {
		return nil, err
	}
	return query(ctx, spark, q)
}

// single runs q, which must return one value.
func single(ctx context.Context, spark sql.SparkSession, tbl, q string) (string, error) {
	rows, err := refreshed(ctx, spark, tbl, q)
	if err != nil {
		return "", err
	}
	if len(rows) != 1 || len(rows[0]) != 1 {
		return "", fmt.Errorf("%s returned %v, want one value", q, rows)
	}
	return rows[0][0], nil
}

// countCheck compares the single value of q with want.
func countCheck(name, desc, q string, want int64) Check {
	return Check{
		Name: fmt.Sprintf("%s %s = %d", name, desc, want),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			got, err := single(ctx, spark, tbl, fmt.Sprintf(q, tbl))
			if err != nil {
				return err
			}
			if got != fmt.Sprint(want) {
				return fmt.Errorf("%s is %s, want %d", desc, got, want)
			}
			return nil
		},
	}
}

// Count checks the number of rows, tombstones included.
func Count(name string, want int64) Check {
	return countCheck(name, "COUNT(*)", "SELECT COUNT(*) FROM %s", want)
}

// CountDistinct checks the number of distinct records.
func CountDistinct(name string, want int64) Check {
	return countCheck(name, "COUNT(DISTINCT _olake_id)", "SELECT COUNT(DISTINCT _olake_id) FROM %s", want)
}

// LiveCount checks the number of rows that are not delete tombstones (_op_type 'd').
func LiveCount(name string, want int64) Check {
	return countCheck(name, "live rows", "SELECT COUNT(*) FROM %s WHERE _op_type != 'd'", want)
}

// Absent checks that the table does not exist.
func Absent(name string) Check {
	return Check{
		Name: name + " absent",
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			// REFRESH drops Spark's cached metadata, which can outlive a dropped table.
			if _, err := spark.Sql(ctx, "REFRESH TABLE "+tbl); isTableMissing(err) {
				return nil
			}
			_, err := query(ctx, spark, "SELECT 1 FROM "+tbl+" LIMIT 1")
			if isTableMissing(err) {
				return nil
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("%s exists", tbl)
		},
	}
}

// CountOrAbsent checks that the table holds want rows, or does not exist when want is 0.
func CountOrAbsent(name string, want int64) Check {
	count := Count(name, want)
	return Check{
		Name: fmt.Sprintf("%s COUNT(*) = %d or absent", name, want),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			err := count.Run(ctx, spark, table)
			if want == 0 && isTableMissing(err) {
				return nil
			}
			return err
		},
	}
}

// Ops checks the number of rows per _op_type: r (full-refresh read), c (insert), u (update) and
// d (delete). Op types missing from want must not occur.
func Ops(name string, want map[string]int64) Check {
	keys := make([]string, 0, len(want))
	for k := range want {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, want[k]))
	}
	wantText := strings.Join(parts, " ")
	return Check{
		Name: fmt.Sprintf("%s ops %s", name, wantText),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			rows, err := refreshed(ctx, spark, tbl, fmt.Sprintf(
				"SELECT _op_type, COUNT(*) FROM %s GROUP BY _op_type ORDER BY _op_type", tbl))
			if err != nil {
				return err
			}
			got := make([]string, 0, len(rows))
			for _, row := range rows {
				got = append(got, fmt.Sprintf("%s=%s", row[0], row[1]))
			}
			if gotText := strings.Join(got, " "); gotText != wantText {
				return fmt.Errorf("ops are %q, want %q", gotText, wantText)
			}
			return nil
		},
	}
}

// LiveIDs checks the set of ids that have a row other than a delete tombstone.
func LiveIDs(name string, ids ...int) Check {
	want := make([]string, 0, len(ids))
	for _, id := range ids {
		want = append(want, fmt.Sprint(id))
	}
	wantText := strings.Join(want, ",")
	return Check{
		Name: fmt.Sprintf("%s live ids {%s}", name, wantText),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			rows, err := refreshed(ctx, spark, tbl, fmt.Sprintf(
				"SELECT DISTINCT CAST(id AS BIGINT) AS id FROM %s WHERE _op_type != 'd' ORDER BY id", tbl))
			if err != nil {
				return err
			}
			got := make([]string, 0, len(rows))
			for _, row := range rows {
				got = append(got, row[0])
			}
			if gotText := strings.Join(got, ","); gotText != wantText {
				return fmt.Errorf("live ids are {%s}, want {%s}", gotText, wantText)
			}
			return nil
		},
	}
}

// OneRowPerID checks that every live id has exactly one live row, as upsert promises.
func OneRowPerID(name string) Check {
	return Check{
		Name: name + " one live row per id",
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			rows, err := refreshed(ctx, spark, tbl, fmt.Sprintf(
				"SELECT CAST(id AS BIGINT), COUNT(*) FROM %s WHERE _op_type != 'd' GROUP BY id HAVING COUNT(*) > 1", tbl))
			if err != nil {
				return err
			}
			if len(rows) > 0 {
				return fmt.Errorf("ids with more than one live row (id, rows): %v", rows)
			}
			return nil
		},
	}
}

// Value checks column of the latest live row of id.
func Value(name string, id int, column, want string) Check {
	return Check{
		Name: fmt.Sprintf("%s id %d %s = %s", name, id, column, want),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			got, err := single(ctx, spark, tbl, fmt.Sprintf(
				"SELECT CAST(%s AS STRING) FROM %s WHERE id = %d AND _op_type != 'd' ORDER BY _olake_timestamp DESC LIMIT 1",
				ident(column), tbl, id))
			if err != nil {
				return err
			}
			if got != want {
				return fmt.Errorf("%s is %q, want %q", column, got, want)
			}
			return nil
		},
	}
}

// Columns checks that the table has every column in present and none in absent.
func Columns(name string, present, absent []string) Check {
	return Check{
		Name: fmt.Sprintf("%s columns %v present, %v absent", name, present, absent),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			rows, err := refreshed(ctx, spark, tbl, "DESCRIBE TABLE "+tbl)
			if err != nil {
				return err
			}
			have := map[string]bool{}
			for _, row := range rows {
				if len(row) > 0 && row[0] != "" && !strings.HasPrefix(row[0], "#") {
					have[row[0]] = true
				}
			}
			for _, c := range present {
				if !have[ident(c)] {
					return fmt.Errorf("column %s is missing", c)
				}
			}
			for _, c := range absent {
				if have[ident(c)] {
					return fmt.Errorf("column %s is present", c)
				}
			}
			return nil
		},
	}
}

// DeleteKind is a kind of Iceberg delete file, as the <table>.delete_files metadata table shows it.
type DeleteKind string

const (
	EqualityDeletes DeleteKind = "equality"        // content = 2
	DeletionVectors DeleteKind = "deletion vector" // content = 1, PUFFIN
)

// DeleteFiles checks that the table has at least one delete file of kind.
func DeleteFiles(name string, kind DeleteKind) Check {
	where := map[DeleteKind]string{
		EqualityDeletes: "content = 2",
		DeletionVectors: "content = 1 AND upper(file_format) = 'PUFFIN'",
	}[kind]
	return Check{
		Name: fmt.Sprintf("%s has %s delete files", name, kind),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			got, err := single(ctx, spark, tbl, fmt.Sprintf("SELECT COUNT(*) FROM %s.delete_files WHERE %s", tbl, where))
			if err != nil {
				return err
			}
			if got == "0" {
				all, _ := query(ctx, spark, fmt.Sprintf("SELECT content, file_format, COUNT(*) FROM %s.delete_files GROUP BY content, file_format", tbl))
				return fmt.Errorf("no %s delete files; delete files (content, format, count): %v", kind, all)
			}
			return nil
		},
	}
}

// FormatVersion checks the table's Iceberg format version.
func FormatVersion(name string, want int) Check {
	return Check{
		Name: fmt.Sprintf("%s format-version = %d", name, want),
		Run: func(ctx context.Context, spark sql.SparkSession, table func(string) string) error {
			tbl := table(ident(name))
			rows, err := refreshed(ctx, spark, tbl, fmt.Sprintf("SHOW TBLPROPERTIES %s ('format-version')", tbl))
			if err != nil {
				return err
			}
			if len(rows) != 1 || len(rows[0]) != 2 || rows[0][1] != fmt.Sprint(want) {
				return fmt.Errorf("format-version is %v, want %d", rows, want)
			}
			return nil
		},
	}
}

// keyColumn is the primary key of the e2e tables; the source comparison matches rows on it.
const keyColumn = "id"

// diffRows compares the source rows with the destination rows, both with the key in the first
// column, and describes every difference: rows only on one side and rows whose values differ.
func diffRows(columns []string, source, destination [][]string) error {
	index := func(rows [][]string) map[string][]string {
		m := make(map[string][]string, len(rows))
		for _, row := range rows {
			m[row[0]] = row
		}
		return m
	}
	src, dst := index(source), index(destination)
	keys := make([]string, 0, len(src)+len(dst))
	for k := range src {
		keys = append(keys, k)
	}
	for k := range dst {
		if _, ok := src[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		a, errA := strconv.ParseInt(keys[i], 10, 64)
		b, errB := strconv.ParseInt(keys[j], 10, 64)
		if errA == nil && errB == nil {
			return a < b
		}
		return keys[i] < keys[j]
	})

	var diffs []string
	for _, k := range keys {
		s, inSrc := src[k]
		d, inDst := dst[k]
		switch {
		case !inDst:
			diffs = append(diffs, fmt.Sprintf("%s %s: only in the source %v", keyColumn, k, s))
		case !inSrc:
			diffs = append(diffs, fmt.Sprintf("%s %s: only in the destination %v", keyColumn, k, d))
		default:
			for i := range columns {
				if s[i] != d[i] {
					diffs = append(diffs, fmt.Sprintf("%s %s: %s is %q in the source, %q in the destination", keyColumn, k, columns[i], s[i], d[i]))
				}
			}
		}
	}
	if len(source) != len(src) || len(destination) != len(dst) {
		diffs = append(diffs, fmt.Sprintf("duplicate %s values: %d source rows for %d ids, %d destination rows for %d ids",
			keyColumn, len(source), len(src), len(destination), len(dst)))
	}
	if len(diffs) > 0 {
		return fmt.Errorf("%d difference(s) between source and destination:\n  %s", len(diffs), strings.Join(diffs, "\n  "))
	}
	return nil
}
