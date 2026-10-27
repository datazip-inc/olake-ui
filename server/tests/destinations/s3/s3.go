// Package s3 is the Amazon S3 destination writing parquet files, with MinIO standing in for S3.
package s3

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
)

// Bucket and Path are where the job writes, as the spec configures the destination:
// s3://Bucket/Path/<folder>/<table>/.
const (
	Bucket = "warehouse"
	Path   = "e2e"
)

// Destination starts MinIO and Spark, which reads the parquet files back over s3a (see
// docker-compose.yml for the s3a settings). Ready waits until Spark Connect listens on 15002.
var Destination = destinations.Destination{
	Name:       "s3",
	ComposeDir: "destinations/s3",
	Services:   []string{"minio", "mc", "postgres", "spark"},
	ReadyCmd: `
        for i in $(seq 1 100); do
            if docker exec s3-spark bash -c 'exec 3<>/dev/tcp/127.0.0.1/15002' 2>/dev/null; then
                echo "Spark Connect is listening on 15002."
                exit 0
            fi
            if docker logs s3-spark 2>&1 | grep -qE "failed to launch: .*SparkConnectServer|failed to fetch the Spark Connect jar"; then
                break
            fi
            sleep 3
        done
        echo "Spark Connect did not start on 15002" >&2
        docker logs --tail 20 s3-spark >&2
        docker exec s3-spark sh -c 'tail -n 40 /opt/spark/logs/*SparkConnectServer*.out' >&2
        exit 1
    `,
	Table: func(folder, table string) string {
		// The table's folder also holds OLake's metadata, so read only the parquet files: in the
		// folder itself, or in the partition folders a partition regex creates, up to two deep.
		return fmt.Sprintf("parquet.`s3a://%s/%s/%s/%s/{*.parquet,*/*.parquet,*/*/*.parquet}`", Bucket, Path, folder, table)
	},
}
