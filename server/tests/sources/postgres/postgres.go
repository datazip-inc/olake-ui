// Package postgres is the Postgres source: wal2json logical replication, published on port 5433.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // registers the "postgres" driver sqlx connects with
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

// Source starts Postgres from this folder's docker-compose.yml and seeds the e2e Tables.
var Source = sources.Source{
	Name:       "postgres",
	ComposeDir: "sources/postgres",
	Port:       "5433",
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
		var statements []string
		for _, table := range Tables {
			statements = append(statements, seedStatements(table)...)
		}
		statements = append(statements, typesTable...)
		statements = append(statements, noPKTable...)
		for _, statement := range statements {
			_, err := db.ExecContext(ctx, statement)
			require.NoError(t, err, "failed to seed: %s", statement)
		}
	},
	Exec:  Exec,
	Query: Query,
}

// Tables are the e2e streams of the integration job, one per feature, that start with rows 1-5:
// id, name 'name_<id>' and amount <id>*10. e2e_types and e2e_no_pk have columns of their own.
var Tables = []string{
	"e2e_cdc",
	"e2e_strict_cdc",
	"e2e_upsert_eq",
	"e2e_upsert_dv",
	"e2e_filter",
	"e2e_new_columns",
	"e2e_fr_append",
	"e2e_fr_upsert",
	"e2e_schema_change",
	"e2e_dropped",
}

// seedStatements recreates table with rows 1-5.
func seedStatements(table string) []string {
	return []string{
		fmt.Sprintf("DROP TABLE IF EXISTS %s", table),
		fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY, name TEXT, amount INT)", table),
		fmt.Sprintf("INSERT INTO %s SELECT g, 'name_' || g, g * 10 FROM generate_series(1, 5) AS g", table),
	}
}

// typesTable holds a value of each type the workflow checks row by row: dates, timestamps with
// and without a time zone, decimals, JSON, arrays and unicode text, and a row of NULLs.
var typesTable = []string{
	"DROP TABLE IF EXISTS e2e_types",
	`CREATE TABLE e2e_types (
		id INT PRIMARY KEY,
		d DATE,
		ts TIMESTAMP,
		tstz TIMESTAMPTZ,
		num NUMERIC(12, 2),
		js JSONB,
		arr INT[],
		txt TEXT
	)`,
	`INSERT INTO e2e_types VALUES
		(1, '2024-01-15', '2024-01-15 10:30:45', '2024-01-15 10:30:45+00', 12345.67, '{"k": "v", "n": 1}', '{1,2,3}', 'héllo wörld ✓ 日本語'),
		(2, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
		(3, '1999-12-31', '1999-12-31 23:59:59', '2000-01-01 05:30:00+05:30', -0.01, '[1, "a", null]', '{}', '')`,
}

// noPKTable is e2e_no_pk: the base columns without a primary key. Its replica identity is FULL:
// the publication covers all tables, and Postgres refuses an UPDATE or DELETE on a published
// table without one.
var noPKTable = []string{
	"DROP TABLE IF EXISTS e2e_no_pk",
	"CREATE TABLE e2e_no_pk (id INT, name TEXT, amount INT)",
	"ALTER TABLE e2e_no_pk REPLICA IDENTITY FULL",
	"INSERT INTO e2e_no_pk SELECT g, 'name_' || g, g * 10 FROM generate_series(1, 5) AS g",
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

// Query runs query on the source Postgres and returns every row as text, NULL as "NULL".
func Query(ctx context.Context, host, port, query string) ([][]string, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", connString(host, port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", query, err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out [][]string
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			return nil, err
		}
		row := make([]string, len(columns))
		for i, v := range values {
			row[i] = "NULL"
			if v.Valid {
				row[i] = v.String
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func connString(host, port string) string {
	return fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port)
}

func connect(ctx context.Context, t *testing.T, host, port string) *sqlx.DB {
	t.Helper()
	db, err := sqlx.ConnectContext(ctx, "postgres", connString(host, port))
	require.NoError(t, err, "failed to connect to postgres")
	return db
}
