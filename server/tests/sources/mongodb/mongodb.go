// Package mongodb is the MongoDB source, seeded by the init script in this folder.
package mongodb

import "github.com/datazip-inc/olake-ui/server/tests/sources"

// The collections of init.js, one per behaviour the spec and the checks look at.
const (
	CollectionFullRefresh = "e2e_full_refresh" // full refresh, normalization on: fields become columns
	CollectionNormOff     = "e2e_norm_off"     // the same documents, normalization off: one data column
	CollectionIncremental = "e2e_incremental"  // cursor updated_at with fallback created_at
	CollectionPartition   = "e2e_partition"    // parquet partition regex on category
	CollectionAdded       = "e2e_added"        // added to the job through Edit Streams
)

// Source starts MongoDB from this folder's docker-compose.yml. init.js seeds it on first start, so there is no Seed.
var Source = sources.Source{
	Name:       "mongodb",
	ComposeDir: "sources/mongodb",
	Port:       "27017/tcp",
	ReadyCmd: `
        for i in $(seq 1 30); do
            if docker exec olake_mongodb-test mongosh -u admin -p password --authenticationDatabase admin --quiet \
                --eval "db.getSiblingDB('test_db').e2e_added.countDocuments({}) > 0 ? 0 : quit(1)"; then
                echo "MongoDB ready."
                break
            fi
            sleep 2
        done
    `,
}
