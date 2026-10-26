package tests

import (
	"context"
	"errors"
	"os"
	"testing"

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

func TestRunWorkflowRejectsUnsafeInput(t *testing.T) {
	for _, w := range []Workflow{
		{Job: "job'; id; '", Spec: "tests/a.spec.ts"},
		{Job: "job", Spec: "tests/a.spec.ts; id"},
		{Job: "job", Spec: "-c/tmp/evil.config.ts"},
		{Job: "job", Spec: "tests/flows/notaspec.ts"},
	} {
		require.Error(t, RunWorkflow(t, w))
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
