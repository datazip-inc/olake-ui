// Package postgres is the Postgres source: wal2json logical replication, published on port 5433.
package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

// Source starts Postgres from this folder's docker-compose.yml and seeds the e2e Tables.
var Source = sources.Source{
	Name:       "postgres",
	ComposeDir: "sources/postgres",
	Port:       "5433/tcp",
	Namespace:  "postgres_public",
	Container:  "olake_postgres-test",
	// Ready once the pgoutput slot olake_slot and the publication olake_publication exist: the
	// compose command creates them after Postgres accepts connections.
	ReadyCmd: `
        for i in $(seq 1 60); do
            ready=$(docker exec olake_postgres-test psql -h localhost -U postgres -d postgres -tAc \
                "SELECT (SELECT count(*) FROM pg_replication_slots WHERE slot_name = 'olake_slot' AND plugin = 'pgoutput') + (SELECT count(*) FROM pg_publication WHERE pubname = 'olake_publication')" 2>/dev/null)
            if [ "$ready" = "2" ]; then
                echo "PostgreSQL ready: olake_slot (pgoutput) and olake_publication exist."
                exit 0
            fi
            sleep 2
        done
        echo "PostgreSQL, olake_slot or olake_publication not ready after 120s" >&2
        docker logs --tail 50 olake_postgres-test >&2
        exit 1
    `,
	Seed: func(ctx context.Context, t *testing.T, host, port string) {
		t.Helper()
		db := connect(ctx, t, host, port)
		defer func() {
			require.NoError(t, db.Close(), "failed to close postgres connection")
		}()
		for _, table := range Tables {
			for _, statement := range seedStatements(table) {
				_, err := db.ExecContext(ctx, statement)
				require.NoError(t, err, "failed to seed %s: %s", table, statement)
			}
		}
	},
	Exec: Exec,
}

// Tables are the e2e streams of the integration job, one per feature. Each starts with rows
// 1-5: id, name 'name_<id>' and amount <id>*10.
var Tables = []string{
	"e2e_cdc",
	"e2e_strict_cdc",
	"e2e_upsert_eq",
	"e2e_upsert_dv",
	"e2e_filter",
	"e2e_new_columns",
}

// seedStatements recreates table with rows 1-5.
func seedStatements(table string) []string {
	return []string{
		fmt.Sprintf("DROP TABLE IF EXISTS %s", table),
		fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY, name TEXT, amount INT)", table),
		fmt.Sprintf("INSERT INTO %s SELECT g, 'name_' || g, g * 10 FROM generate_series(1, 5) AS g", table),
	}
}

// Exec runs statements on the source Postgres in order, each in its own transaction.
func Exec(ctx context.Context, t *testing.T, host, port string, statements ...string) {
	t.Helper()
	db := connect(ctx, t, host, port)
	defer func() {
		require.NoError(t, db.Close(), "failed to close postgres connection")
	}()
	for _, statement := range statements {
		t.Logf("source SQL: %s", statement)
		_, err := db.ExecContext(ctx, statement)
		require.NoError(t, err, "failed to run %s", statement)
	}
}

func connect(ctx context.Context, t *testing.T, host, port string) *sqlx.DB {
	t.Helper()
	connStr := fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port)
	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	require.NoError(t, err, "failed to connect to postgres")
	return db
}
