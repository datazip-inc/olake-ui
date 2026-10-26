package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServiceHost(t *testing.T) {
	detect := func(host string, err error) func(context.Context) (string, error) {
		return func(context.Context) (string, error) { return host, err }
	}
	ctx := context.Background()

	t.Run("detected address", func(t *testing.T) {
		t.Setenv("E2E_SERVICE_HOST", "")
		host, err := serviceHost(ctx, detect("172.17.0.1\n", nil))
		require.NoError(t, err)
		require.Equal(t, "172.17.0.1", host)
	})
	t.Run("override wins", func(t *testing.T) {
		t.Setenv("E2E_SERVICE_HOST", "host.docker.internal")
		host, err := serviceHost(ctx, detect("", errors.New("not called")))
		require.NoError(t, err)
		require.Equal(t, "host.docker.internal", host)
	})
	t.Run("detect error", func(t *testing.T) {
		t.Setenv("E2E_SERVICE_HOST", "")
		_, err := serviceHost(ctx, detect("", errors.New("no bridge")))
		require.Error(t, err)
	})
	t.Run("empty detection names the override", func(t *testing.T) {
		t.Setenv("E2E_SERVICE_HOST", "")
		_, err := serviceHost(ctx, detect("", nil))
		require.ErrorContains(t, err, "E2E_SERVICE_HOST")
	})
	for _, bad := range []string{"a b", "1.2.3.4;rm -rf /", "'x'", "$(id)", "-x", "fd00::1", "1.2.3.4\nx"} {
		t.Run("rejects "+bad, func(t *testing.T) {
			t.Setenv("E2E_SERVICE_HOST", "")
			_, err := serviceHost(ctx, detect(bad, nil))
			require.Error(t, err)
		})
	}
}

func TestPlaywrightCmd(t *testing.T) {
	cmd, err := playwrightCmd("/mnt", "tests/integration/postgres.spec.ts", "@first-sync",
		map[string]string{"E2E_JOB_NAME": "job", "E2E_WEBHOOK_URL": "http://172.17.0.2:8099/alert"}, false)
	require.NoError(t, err)
	require.Contains(t, cmd, "cd '/mnt/ui'")
	require.Contains(t, cmd, "E2E_JOB_NAME='job' E2E_WEBHOOK_URL='http://172.17.0.2:8099/alert' ")
	require.Contains(t, cmd, "npx playwright test 'tests/integration/postgres.spec.ts' --grep '@first-sync(?![a-z0-9-])'")
	require.NotContains(t, cmd, "DEBUG=")

	debugCmd, err := playwrightCmd("/mnt", "tests/integration/postgres.spec.ts", "", nil, true)
	require.NoError(t, err)
	require.Contains(t, debugCmd, "DEBUG=pw:api PLAYWRIGHT_TEST_BASE_URL=")

	for _, c := range []struct{ spec, tag, key string }{
		{"tests/a.spec.ts; id", "", "E2E_X"},
		{"-c/tmp/evil.config.ts", "", "E2E_X"},
		{"tests/flows/notaspec.ts", "", "E2E_X"},
		{"tests/a.spec.ts", "sync", "E2E_X"},
		{"tests/a.spec.ts", "@sync; id", "E2E_X"},
		{"tests/a.spec.ts", "", "PATH"},
		{"tests/a.spec.ts", "", "E2E_X; id"},
	} {
		_, err := playwrightCmd("/mnt", c.spec, c.tag, map[string]string{c.key: "v"}, false)
		require.Error(t, err, "%+v", c)
	}
}

func TestRunnerEnvOnlyOnGitHubHosted(t *testing.T) {
	t.Setenv("OLAKE_TEST_ENV", "runner")
	t.Setenv("OLAKE_TEST_RUNNER_ANY_HOST", "")
	for _, env := range []string{"", "self-hosted"} {
		t.Setenv("RUNNER_ENVIRONMENT", env)
		_, err := newEnv(context.Background(), t, nil)
		require.ErrorContains(t, err, "GitHub-hosted")
	}
}

func TestShellQuote(t *testing.T) {
	dir := t.TempDir() + "/a dir with 'quote' $HOME;x"
	require.NoError(t, os.MkdirAll(dir, 0o755))
	r := &runnerEnv{root: dir}
	code, out, err := r.Exec(context.Background(), "cd "+shellQuote(dir)+" && pwd")
	require.NoError(t, err)
	require.Equal(t, 0, code, string(out))
	require.Equal(t, dir+"\n", string(out))
}

func TestClearDirCmd(t *testing.T) {
	cmd := clearDirCmd("/mnt/server/tests/a dir")
	require.Equal(t, "mkdir -p '/mnt/server/tests/a dir' && docker run --rm -v '/mnt/server/tests/a dir':/clear alpine:3.20 find /clear -mindepth 1 -delete", cmd)
}

func TestTasksRunInParallelAndCancelOnFailure(t *testing.T) {
	r := &Run{t: t, env: &runnerEnv{root: "/tmp"}, start: time.Now()}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Three one-second tasks take about one second together.
	start := time.Now()
	var chans []<-chan error
	for i := 0; i < 3; i++ {
		chans = append(chans, r.task(ctx, cancel, &wg, fmt.Sprint("sleep ", i), func(ctx context.Context) error {
			return r.run(ctx, "sleep 1", "sleep failed")
		}))
	}
	for _, c := range chans {
		require.NoError(t, <-c)
	}
	require.Less(t, time.Since(start), 2500*time.Millisecond, "tasks ran one after another")

	// A failing task cancels a long one.
	start = time.Now()
	long := r.task(ctx, cancel, &wg, "long", func(ctx context.Context) error {
		return r.run(ctx, "sleep 30", "long task failed")
	})
	failing := r.task(ctx, cancel, &wg, "failing", func(ctx context.Context) error {
		return r.run(ctx, "exit 3", "failing task")
	})
	require.ErrorContains(t, <-failing, "failing task (3)")
	require.Error(t, <-long)
	wg.Wait()
	require.Less(t, time.Since(start), 5*time.Second, "the long task was not cancelled")
}
