// Package jdbc is the Apache Iceberg destination with a JDBC catalog, data in MinIO.
package jdbc

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
)

// Destination starts MinIO, the catalog Postgres and Spark; Spark reads the tables through olake_iceberg.
//
// The image does not ship the Spark Connect jar: start-connect-server.sh downloads it with
// --packages on every start, from Maven Central or, where Central blocks the network, from
// Google's mirror of it (spark.jars.repositories in docker-compose.yml). Ready waits until Spark
// Connect listens on 15002, so a failed start shows here with its log, not at the first check.
var Destination = destinations.Destination{
	Name:       "jdbc",
	ComposeDir: "destinations/jdbc",
	Services:   []string{"minio", "mc", "postgres", "spark-iceberg"},
	ReadyCmd: `
        for i in $(seq 1 100); do
            if docker exec spark-iceberg bash -c 'exec 3<>/dev/tcp/127.0.0.1/15002' 2>/dev/null; then
                echo "Spark Connect is listening on 15002."
                exit 0
            fi
            if docker logs spark-iceberg 2>&1 | grep -q "failed to launch: .*SparkConnectServer"; then
                break
            fi
            sleep 3
        done
        echo "Spark Connect did not start on 15002" >&2
        docker exec spark-iceberg sh -c 'tail -n 40 /opt/spark/logs/*SparkConnectServer*.out' >&2
        exit 1
    `,
	DataDir: "data", // ./data/minio-data and ./data/postgres-data (the Iceberg catalog)
	Table: func(database, table string) string {
		return fmt.Sprintf("olake_iceberg.%s.%s", database, table)
	},
}
