package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/stretchr/testify/require"
)

// CountDistinct checks that the table holds want distinct records.
func CountDistinct(name string, want int64) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT _olake_id) as unique_count FROM %s", table(name))
		countValue := queryCount(ctx, t, spark, countQuery)
		require.Equal(t, want, countValue, "Expected count to be %d", want)
		t.Logf("✅ Test passed: count value %v matches expected value %d", countValue, want)
	}
}

// Count checks that the table holds want rows. Unlike CountDistinct it fails when a record
// was written twice.
func Count(name string, want int64) Check {
	return func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(string) string) {
		countQuery := fmt.Sprintf("SELECT COUNT(*) as unique_count FROM %s", table(name))
		countValue := queryCount(ctx, t, spark, countQuery)
		require.Equal(t, want, countValue, "Expected row count to be %d", want)
		t.Logf("✅ Test passed: row count %v matches expected value %d", countValue, want)
	}
}

// queryCount runs a query that returns one unique_count column and returns its value.
func queryCount(ctx context.Context, t *testing.T, spark sql.SparkSession, countQuery string) int64 {
	t.Helper()
	t.Logf("Executing query: %s", countQuery)

	countQueryDf, err := spark.Sql(ctx, countQuery)
	require.NoError(t, err, "Failed to execute query on the table")

	rows, err := countQueryDf.Collect(ctx)
	require.NoError(t, err, "Failed to collect data rows from the destination")
	require.NotEmpty(t, rows, "No rows returned for the count query")

	value := rows[0].Value("unique_count")
	countValue, ok := value.(int64)
	require.True(t, ok, "unique_count is %T (%v), want int64", value, value)
	return countValue
}
