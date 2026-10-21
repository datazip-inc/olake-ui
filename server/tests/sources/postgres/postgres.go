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
	// Ready once olake_slot exists: the compose command creates it after Postgres accepts
	// connections, so a plain "SELECT 1" can pass before the slot is there.
	ReadyCmd: `
        for i in $(seq 1 60); do
            slots=$(docker exec olake_postgres-test psql -h localhost -U postgres -d postgres -tAc \
                "SELECT count(*) FROM pg_replication_slots WHERE slot_name = 'olake_slot'" 2>/dev/null)
            if [ "$slots" = "1" ]; then
                echo "PostgreSQL ready, olake_slot exists."
                exit 0
            fi
            sleep 2
        done
        echo "PostgreSQL or olake_slot not ready after 120s" >&2
        docker logs --tail 50 olake_postgres-test >&2
        exit 1
    `,
	Seed: func(ctx context.Context, t *testing.T, host, port string) {
		ExecuteQuery(ctx, t, "create", host, port)
		ExecuteQuery(ctx, t, "clean", host, port)
		ExecuteQuery(ctx, t, "add", host, port)
	},
}

func ExecuteQuery(ctx context.Context, t *testing.T, operation, host, port string) {
	t.Helper()
	connStr := fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port)
	db, ok := sqlx.ConnectContext(ctx, "postgres", connStr)
	require.NoError(t, ok, "failed to connect to postgres")
	defer func() {
		require.NoError(t, db.Close(), "failed to close postgres connection")
	}()

	// integration test uses only one stream for testing
	integrationTestTable := TestTable
	var query string

	switch operation {
	case "create":
		query = fmt.Sprintf(`
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
			)`, integrationTestTable)

	case "drop":
		query = fmt.Sprintf("DROP TABLE IF EXISTS %s", integrationTestTable)

	case "clean":
		query = fmt.Sprintf("TRUNCATE TABLE %s", integrationTestTable)

	case "add":
		insertTestData(ctx, t, db, integrationTestTable)
		return // Early return since we handle all inserts in the helper function

	case "insert":
		query = fmt.Sprintf(`
			INSERT INTO %s (
				col_bigint, col_bool, col_char, col_character,
				col_character_varying, col_date, col_decimal,
				col_double_precision, col_float4, col_int, col_int2,
				col_integer, col_interval, col_json, col_jsonb,
				col_name, col_numeric, col_real, col_text,
				col_timestamp, col_timestamptz, col_uuid, col_varbit, col_xml
			) VALUES (
				123456789012345, TRUE, 'c', 'char_val',
				'varchar_val', '2023-01-01', 123.45,
				123.456789, 123.45, 123, 123, 12345,
				'1 hour', '{"key": "value"}', '{"key": "value"}',
				'test_name', 123.45, 123.45, 'sample text',
				'2023-01-01 12:00:00', '2023-01-01 12:00:00+00',
				'123e4567-e89b-12d3-a456-426614174000', B'101010',
				'<tag>value</tag>'
			)`, integrationTestTable)

	case "update":
		query = fmt.Sprintf(`
			UPDATE %s SET
				col_bigint = 123456789012340,
				col_bool = FALSE,
				col_char = 'd',
				col_character = 'updated__',
				col_character_varying = 'updated val',
				col_date = '2024-07-01',
				col_decimal = 543.21,
				col_double_precision = 987.654321,
				col_float4 = 543.21,
				col_int = 321,
				col_int2 = 321,
				col_integer = 54321,
				col_interval = '2 hours',
				col_json = '{"new": "json"}',
				col_jsonb = '{"new": "jsonb"}',
				col_name = 'updated_name',
				col_numeric = 321.00,
				col_real = 321.00,
				col_text = 'updated text',
				col_timestamp = '2024-07-01 15:30:00',
				col_timestamptz = '2024-07-01 15:30:00+00',
				col_uuid = '00000000-0000-0000-0000-000000000000',
				col_varbit = B'111000',
				col_xml = '<updated>value</updated>'
			WHERE col_bigserial = 1`, integrationTestTable)

	case "delete":
		query = fmt.Sprintf("DELETE FROM %s WHERE col_bigserial = 1", integrationTestTable)

	default:
		t.Fatalf("Unsupported operation: %s", operation)
	}
	_, err := db.ExecContext(ctx, query)
	require.NoError(t, err, "Failed to execute %s operation", operation)
}

// insertTestData inserts test data into the specified table
func insertTestData(ctx context.Context, t *testing.T, db *sqlx.DB, tableName string) {
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
		)`, tableName)

		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err, "Failed to insert test data")
	}
}
