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
		t.Logf("Executing query: %s", countQuery)

		countQueryDf, err := spark.Sql(ctx, countQuery)
		require.NoError(t, err, "Failed to execute query on the table")

		rows, err := countQueryDf.Collect(ctx)
		require.NoError(t, err, "Failed to collect data rows from the destination")
		require.NotEmpty(t, rows, "No rows returned for the count query")

		countValue := rows[0].Value("unique_count").(int64)
		require.Equal(t, want, countValue, "Expected count to be %d", want)
		t.Logf("✅ Test passed: count value %v matches expected value %d", countValue, want)
	}
}
