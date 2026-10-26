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

// TestTable is the stream the integration job syncs.
const TestTable = "postgres_test_table_olake"

// Source starts Postgres from this folder's docker-compose.yml and seeds TestTable with 5 rows.
var Source = sources.Source{
	Name:       "postgres",
	ComposeDir: "sources/postgres",
	Port:       "5433/tcp",
	Namespace:  "postgres_public",
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
		createTable(ctx, t, db)
		truncateTable(ctx, t, db)
		insertTestData(ctx, t, db)
	},
}

func connect(ctx context.Context, t *testing.T, host, port string) *sqlx.DB {
	t.Helper()
	connStr := fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port)
	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	require.NoError(t, err, "failed to connect to postgres")
	return db
}

func createTable(ctx context.Context, t *testing.T, db *sqlx.DB) {
	t.Helper()
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			col_bigint BIGINT,
			col_bigserial BIGSERIAL PRIMARY KEY,
			col_bool BOOLEAN,
			col_char CHAR(1),
			col_character CHAR(10),
			col_character_varying VARCHAR(50),
			col_date DATE,
			col_decimal NUMERIC,
			col_double_precision DOUBLE PRECISION,
			col_float4 REAL,
			col_int INT,
			col_int2 SMALLINT,
			col_integer INTEGER,
			col_interval INTERVAL,
			col_json JSON,
			col_jsonb JSONB,
			col_name NAME,
			col_numeric NUMERIC,
			col_real REAL,
			col_text TEXT,
			col_timestamp TIMESTAMP,
			col_timestamptz TIMESTAMPTZ,
			col_uuid UUID,
			col_varbit VARBIT(20),
			col_xml XML,
			CONSTRAINT unique_custom_key UNIQUE (col_bigserial)
		)`, TestTable)
	_, err := db.ExecContext(ctx, query)
	require.NoError(t, err, "Failed to create %s", TestTable)
}

func truncateTable(ctx context.Context, t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", TestTable))
	require.NoError(t, err, "Failed to truncate %s", TestTable)
}

// insertTestData inserts 5 rows into TestTable
func insertTestData(ctx context.Context, t *testing.T, db *sqlx.DB) {
	t.Helper()

	for i := 1; i <= 5; i++ {
		query := fmt.Sprintf(`
		INSERT INTO %s (
			col_bigint, col_bigserial, col_bool, col_char, col_character,
			col_character_varying, col_date, col_decimal,
			col_double_precision, col_float4, col_int, col_int2, col_integer,
			col_interval, col_json, col_jsonb, col_name, col_numeric,
			col_real, col_text, col_timestamp, col_timestamptz,
			col_uuid, col_varbit, col_xml
		) VALUES (
			123456789012345, DEFAULT, TRUE, 'c', 'char_val',
			'varchar_val', '2023-01-01', 123.45,
			123.456789, 123.45, 123, 123, 12345, '1 hour', '{"key": "value"}',
			'{"key": "value"}', 'test_name', 123.45, 123.45,
			'sample text', '2023-01-01 12:00:00',
			'2023-01-01 12:00:00+00',
			'123e4567-e89b-12d3-a456-426614174000', B'101010',
			'<tag>value</tag>'
		)`, TestTable)

		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err, "Failed to insert test data")
	}
}
