package optimization

import (
	"strings"
	"testing"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

func TestCreateAlterQuery_SingleStatement(t *testing.T) {
	db, tbl := "mydb", "orders"
	props := map[string]string{
		constants.OptEnableOptimization: "true",
		constants.OptMinorCron:          "0 * * * *",
	}
	sql := createAlterQuery(db, tbl, props)
	if !strings.HasPrefix(sql, "ALTER TABLE mydb.orders SET TBLPROPERTIES (") {
		t.Fatalf("unexpected prefix: %q", sql)
	}
	if !strings.HasSuffix(strings.TrimSpace(sql), ";") {
		t.Fatalf("expected trailing semicolon: %q", sql)
	}
}

func TestBulkAlterScript_OrderAndStatements(t *testing.T) {
	db := "mydb"
	tableNames := []string{"zebra", "alpha"}
	props := map[string]string{
		constants.OptEnableOptimization: "true",
		constants.OptMinorCron:          "0 * * * *",
	}
	stmts := make([]string, 0, len(tableNames))
	for _, t := range tableNames {
		stmts = append(stmts, createAlterQuery(db, t, props))
	}
	sql := strings.Join(stmts, "\n")

	if !strings.Contains(sql, "ALTER TABLE mydb.zebra SET TBLPROPERTIES") {
		t.Fatalf("missing zebra statement: %q", sql)
	}
	if !strings.Contains(sql, "ALTER TABLE mydb.alpha SET TBLPROPERTIES") {
		t.Fatalf("missing alpha statement: %q", sql)
	}
	zebraPos := strings.Index(sql, "mydb.zebra")
	alphaPos := strings.Index(sql, "mydb.alpha")
	if zebraPos > alphaPos {
		t.Fatalf("expected zebra before alpha in script:\n%s", sql)
	}
	if !strings.HasSuffix(strings.TrimSpace(strings.Split(sql, "\n")[0]), ";") {
		t.Fatalf("expected semicolon-terminated first statement: %q", sql)
	}
}
