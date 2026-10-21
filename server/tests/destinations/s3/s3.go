// Package s3 is the Amazon S3 destination writing parquet files, with MinIO standing in for S3.
package s3

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
)

// Bucket and Path are where the job writes, as configured in the UI spec: s3://Bucket/Path/<folder>/<table>.
const (
	Bucket = "warehouse"
	Path   = "e2e"
)

// Destination starts MinIO and Spark; Spark reads the parquet files back over s3a.
var Destination = destinations.Destination{
	Name:       "s3",
	ComposeDir: "destinations/s3",
	Table: func(folder, table string) string {
		// the table folder also holds a metadata.json, so read only the parquet files, in the
		// folder itself or one level down where a partition regex puts them
		return fmt.Sprintf("parquet.`s3a://%s/%s/%s/%s/{*.parquet,*/*.parquet}`", Bucket, Path, folder, table)
	},
	SparkInit: []string{
		"SET fs.s3a.endpoint=http://minio:9000",
		"SET fs.s3a.path.style.access=true",
		"SET fs.s3a.connection.ssl.enabled=false",
	},
}
