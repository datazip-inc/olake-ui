package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/stretchr/testify/require"
)

func TestDinDIntegration(t *testing.T) {
	err := DinDTestContainer(t)
	if err != nil {
		t.Errorf("Error in Docker in Docker container start up: %s", err)
	}
}

func TestVerification(t *testing.T) {
	ctx := context.Background()
	spark, err := sql.NewSessionBuilder().Remote(sparkConnectAddress).Build(ctx)
	require.NoError(t, err, "Failed to connect to Spark Connect server")
	defer func() {
		if stopErr := spark.Stop(); stopErr != nil {
			t.Errorf("Failed to stop Spark session: %v", stopErr)
		}
		if dindContainer != nil {
			t.Log("Running cleanup...")
			// Stop docker-compose services
			_, _, _ = ExecCommand(ctx, dindContainer, "cd /mnt && docker-compose down -v --remove-orphans")
			// Terminate the DinD container
			if err := dindContainer.Terminate(ctx); err != nil {
				t.Logf("Warning: failed to terminate container: %v", err)
			}
			t.Log("Cleanup complete")
		}
	}()
	countQuery := fmt.Sprintf(
		"SELECT COUNT(DISTINCT _olake_id) as unique_count FROM %s.%s.%s WHERE _op_type = 'r'",
		icebergCatalog, icebergDB, currentTestTable,
	)
	t.Logf("Executing query: %s", countQuery)

	countQueryDf, err := spark.Sql(ctx, countQuery)
	require.NoError(t, err, "Failed to execute query on the table")

	rows, err := countQueryDf.Collect(ctx)
	require.NoError(t, err, "Failed to collect data rows from Iceberg")
	require.NotEmpty(t, rows, "No rows returned for _op_type = 'r'")

	// check count and verify
	countValue := rows[0].Value("unique_count").(int64)
	require.Equal(t, int64(5), countValue, "Expected count to be 5")
}
