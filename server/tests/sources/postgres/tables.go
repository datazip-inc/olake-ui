package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// Feature tables, one per behaviour the Postgres workflow checks. Each has the columns id, name, amount.
const (
	TableCDC        = "e2e_cdc"         // Full Refresh + CDC, Append
	TableStrictCDC  = "e2e_strict_cdc"  // CDC Only
	TableUpsertEq   = "e2e_upsert_eq"   // Upsert with Equality Deletes
	TableUpsertDV   = "e2e_upsert_dv"   // Upsert with Deletion Vector
	TableFilter     = "e2e_filter"      // Data Filter: amount > 20 AND id < 5
	TableNewColumns = "e2e_new_columns" // Sync new columns automatically
	TableFailed     = "e2e_failed"      // its own job, dropped at the source so the sync fails
)

// FeatureTables lists every feature table in the order they are created.
var FeatureTables = []string{TableCDC, TableStrictCDC, TableUpsertEq, TableUpsertDV, TableFilter, TableNewColumns, TableFailed}

// SeedFeatureTables recreates the feature tables with their first rows.
func SeedFeatureTables(ctx context.Context, t *testing.T, host, port string) {
	t.Helper()
	exec(ctx, t, host, port, func(db *sqlx.DB) {
		// the source's CDC check needs the publication to exist
		run(ctx, t, db, "DROP PUBLICATION IF EXISTS olake_publication")
		run(ctx, t, db, "CREATE PUBLICATION olake_publication FOR ALL TABLES")
		for _, table := range FeatureTables {
			run(ctx, t, db, fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
			run(ctx, t, db, fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY, name TEXT, amount INT)", table))
			run(ctx, t, db, fmt.Sprintf("ALTER TABLE %s REPLICA IDENTITY FULL", table))
		}
		for _, table := range []string{TableCDC, TableUpsertEq, TableUpsertDV} {
			run(ctx, t, db, fmt.Sprintf("INSERT INTO %s VALUES (1,'a',10),(2,'b',20),(3,'c',30),(4,'d',40),(5,'e',50)", table))
		}
		// the strict CDC table is empty on purpose: the first sync must write 0 rows
		// only id 3 matches amount > 20 AND id < 5: 1 and 2 fail on amount, 6 fails on id
		run(ctx, t, db, fmt.Sprintf("INSERT INTO %s VALUES (1,'a',10),(2,'b',20),(3,'c',30),(6,'f',60)", TableFilter))
		run(ctx, t, db, fmt.Sprintf("INSERT INTO %s VALUES (1,'a',10),(2,'b',20)", TableNewColumns))
		run(ctx, t, db, fmt.Sprintf("INSERT INTO %s VALUES (1,'a',10)", TableFailed))
	})
}

func exec(ctx context.Context, t *testing.T, host, port string, fn func(db *sqlx.DB)) {
	t.Helper()
	connStr := fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port)
	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	require.NoError(t, err, "failed to connect to postgres")
	defer func() {
		require.NoError(t, db.Close(), "failed to close postgres connection")
	}()
	fn(db)
}

func run(ctx context.Context, t *testing.T, db *sqlx.DB, query string) {
	t.Helper()
	_, err := db.ExecContext(ctx, query)
	require.NoError(t, err, "failed to run %q", query)
}
