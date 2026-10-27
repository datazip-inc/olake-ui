// Package mongodb is the MongoDB source, a standalone server (no replica set, so no CDC) seeded
// by init.js in this folder on its first start.
package mongodb

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

// container is the MongoDB container docker-compose.yml starts.
const container = "olake_mongodb-test"

// Source starts MongoDB from this folder's docker-compose.yml. init.js seeds it on first start,
// so Seed is nil; Ready waits for the last collection init.js creates.
var Source = sources.Source{
	Name:       "mongodb",
	ComposeDir: "sources/mongodb",
	Port:       "27017/tcp",
	Namespace:  "test_db",
	Container:  container,
	ReadyCmd: `
        for i in $(seq 1 60); do
            if docker exec ` + container + ` mongosh -u admin -p password --authenticationDatabase admin --quiet \
                --eval "quit(db.getSiblingDB('test_db').e2e_int_ids.countDocuments({}) > 0 ? 0 : 1)" >/dev/null 2>&1; then
                echo "MongoDB ready: init.js has seeded test_db."
                exit 0
            fi
            sleep 2
        done
        echo "MongoDB or its init.js not ready after 120s" >&2
        docker logs --tail 50 ` + container + ` >&2
        exit 1
    `,
	Exec: Exec,
}

// Exec runs JavaScript statements in mongosh on test_db, in order, each on its own. The statements
// are passed to mongosh as an argument, not through a shell.
func Exec(ctx context.Context, t *testing.T, _, _ string, statements ...string) {
	t.Helper()
	for _, statement := range statements {
		t.Logf("source JS: %s", statement)
		out, err := exec.CommandContext(ctx, "docker", "exec", container,
			"mongosh", "-u", "admin", "-p", "password", "--authenticationDatabase", "admin", "--quiet",
			"--eval", `db = db.getSiblingDB("test_db"); `+statement,
		).CombinedOutput()
		require.NoError(t, err, "failed to run %s:\n%s", statement, strings.TrimSpace(string(out)))
	}
}
