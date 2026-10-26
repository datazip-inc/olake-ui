package tests

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/stretchr/testify/require"
)

func TestRunChecksRetriesUntilPass(t *testing.T) {
	defer func(timeout, interval time.Duration) { checkTimeout, checkInterval = timeout, interval }(checkTimeout, checkInterval)
	checkTimeout, checkInterval = time.Second, 10*time.Millisecond

	calls := 0
	flaky := Check{Name: "flaky", Run: func(context.Context, sql.SparkSession, func(string) string) error {
		calls++
		if calls < 3 {
			return errors.New("not visible yet")
		}
		return nil
	}}
	runChecks(context.Background(), t, nil, func(n string) string { return n }, []Check{flaky})
	require.Equal(t, 3, calls)
}

func TestIsTableMissing(t *testing.T) {
	// Messages as Spark Connect returns them.
	notFound := "[TABLE_OR_VIEW_NOT_FOUND] The table or view `olake_iceberg`.`db`.`t` cannot be found."
	noNamespace := "org.apache.iceberg.exceptions.NoSuchNamespaceException: Namespace does not exist: db"
	require.True(t, isTableMissing(errors.New(notFound)))
	require.True(t, isTableMissing(errors.New(noNamespace)))
	require.False(t, isTableMissing(errors.New("[PARSE_SYNTAX_ERROR] Syntax error")))
	require.False(t, isTableMissing(nil))
}

func TestCheckNames(t *testing.T) {
	require.Equal(t, "e2e_cdc ops c=1 d=1 r=5 u=1", Ops("e2e_cdc", map[string]int64{"r": 5, "c": 1, "u": 1, "d": 1}).Name)
	require.Equal(t, "e2e_filter live ids {2,3,4}", LiveIDs("e2e_filter", 2, 3, 4).Name)
	require.Panics(t, func() { ident("t; DROP TABLE x") })
}

func TestDiffRows(t *testing.T) {
	cols := []string{"id", "name", "amount"}
	same := [][]string{{"1", "a", "10"}, {"2", "NULL", "20"}}
	require.NoError(t, diffRows(cols, same, same))

	err := diffRows(cols,
		[][]string{{"1", "a", "10"}, {"2", "b", "20"}, {"10", "j", "100"}},
		[][]string{{"1", "a", "11"}, {"3", "c", "30"}, {"10", "j", "100"}})
	require.Error(t, err)
	msg := err.Error()
	require.Contains(t, msg, "3 difference(s)")
	require.Contains(t, msg, `id 1: amount is "10" in the source, "11" in the destination`)
	require.Contains(t, msg, "id 2: only in the source [2 b 20]")
	require.Contains(t, msg, "id 3: only in the destination [3 c 30]")
	require.Less(t, strings.Index(msg, "id 2:"), strings.Index(msg, "id 3:"), "ids are reported in numeric order")

	err = diffRows(cols, [][]string{{"1", "a", "10"}}, [][]string{{"1", "a", "10"}, {"1", "a", "10"}})
	require.ErrorContains(t, err, "duplicate id values")
}
