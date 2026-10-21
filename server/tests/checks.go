package tests

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/stretchr/testify/require"
)

// query runs sqlText through Spark and returns every row as its values joined by ":".
func query(ctx context.Context, t *testing.T, spark sql.SparkSession, sqlText string) []string {
	t.Helper()
	t.Logf("Executing query: %s", sqlText)
	rows, err := tryQuery(ctx, spark, sqlText)
	require.NoError(t, err, "Failed to run query: %s", sqlText)
	return rows
}

// tryQuery is query without the failure, so a check can look at the error itself.
func tryQuery(ctx context.Context, spark sql.SparkSession, sqlText string) ([]string, error) {
	df, err := spark.Sql(ctx, sqlText)
	if err != nil {
		return nil, err
	}
	rows, err := df.Collect(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		values := make([]string, 0, row.Len())
		for _, v := range row.Values() {
			values = append(values, fmt.Sprint(v))
		}
		out = append(out, strings.Join(values, ":"))
	}
	return out, nil
}

func isMissingTable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "TABLE_OR_VIEW_NOT_FOUND")
}

// OpTypes checks how many rows of each operation type (r, c, u, d) the table holds.
func OpTypes(name string, want map[string]int) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		got := map[string]int{}
		for _, row := range query(ctx, t, spark, fmt.Sprintf("SELECT _op_type, COUNT(*) FROM %s GROUP BY _op_type", table(name))) {
			op, count, _ := strings.Cut(row, ":")
			n, err := strconv.Atoi(count)
			require.NoError(t, err)
			got[op] = n
		}
		require.Equal(t, want, got, "Unexpected operation types in %s", name)
	}
}

// Rows checks the live id:amount pairs of the table, so a duplicate, a missed update or a missed delete fails it.
// A delete is written as a row of operation type d, which is not live. With no want the table has no rows.
func Rows(name string, want ...string) Check {
	if want == nil {
		want = []string{}
	}
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		got := query(ctx, t, spark, fmt.Sprintf("SELECT id, amount FROM %s WHERE _op_type <> 'd' ORDER BY id", table(name)))
		require.Equal(t, want, got, "Unexpected rows in %s", name)
	}
}

// Dropped checks that the table no longer exists, as after Clear Destination.
func Dropped(name string) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		_, err := tryQuery(ctx, spark, fmt.Sprintf("SELECT * FROM %s", table(name)))
		require.True(t, isMissingTable(err), "Expected %s to be dropped, got: %v", name, err)
	}
}

// NewColumn checks that the column added at the source reached the destination with the value of row id 3.
func NewColumn(name, column string) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		got := query(ctx, t, spark, fmt.Sprintf("SELECT %s FROM %s WHERE id = 3", column, table(name)))
		require.Equal(t, []string{"new"}, got, "Expected %s.%s to be synced", name, column)
	}
}

// DeletionVector checks that the table is Iceberg format v3 and that its deletes are deletion vectors.
func DeletionVector(name string) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		version := query(ctx, t, spark, fmt.Sprintf("SHOW TBLPROPERTIES %s ('format-version')", table(name)))
		require.Equal(t, []string{"format-version:3"}, version, "Expected %s to be Iceberg format v3", name)

		formats := query(ctx, t, spark, fmt.Sprintf("SELECT DISTINCT file_format FROM %s.delete_files", table(name)))
		require.Equal(t, []string{"PUFFIN"}, formats, "Expected %s deletes to be deletion vectors", name)
	}
}

// Query runs a statement where %s stands for the table and checks its rows, each as its values joined by ":".
func Query(name, statement string, want ...string) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		got := query(ctx, t, spark, fmt.Sprintf(statement, table(name)))
		require.Equal(t, want, got, "Unexpected rows in %s", name)
	}
}

// DataColumns checks the columns of the table that are not OLake's own (_olake_*, _cdc_timestamp, _op_type).
func DataColumns(name string, want ...string) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		statement := fmt.Sprintf("SELECT * FROM %s LIMIT 1", table(name))
		t.Logf("Executing query: %s", statement)
		df, err := spark.Sql(ctx, statement)
		require.NoError(t, err, "Failed to execute query: %s", statement)
		rows, err := df.Collect(ctx)
		require.NoError(t, err, "Failed to collect rows of: %s", statement)
		require.NotEmpty(t, rows, "Expected %s to hold rows", name)

		var got []string
		for _, column := range rows[0].FieldNames() {
			if !strings.HasPrefix(column, "_olake_") && column != "_cdc_timestamp" && column != "_op_type" {
				got = append(got, column)
			}
		}
		sort.Strings(got)
		require.Equal(t, want, got, "Unexpected columns in %s", name)
	}
}
