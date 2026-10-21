// Package postgres is the Postgres source: wal2json logical replication, published on port 5433.
package postgres

import (
	_ "github.com/lib/pq"

	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

// Source starts Postgres from this folder's docker-compose.yml and seeds the feature tables.
var Source = sources.Source{
	Name:       "postgres",
	ComposeDir: "sources/postgres",
	Port:       "5433/tcp",
	ReadyCmd: `
        for i in $(seq 1 30); do
            if docker exec olake_postgres-test psql -h localhost -U postgres -d postgres -c "SELECT 1" 2>/dev/null; then
                echo "PostgreSQL ready."
                break
            fi
            sleep 2
        done &&
        docker exec olake_postgres-test psql -U postgres -d postgres -c \
            "SELECT slot_name, plugin, slot_type, active FROM pg_replication_slots WHERE slot_name = 'olake_slot';"
    `,
	Seed: SeedFeatureTables,
}
