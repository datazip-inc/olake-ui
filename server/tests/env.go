package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// Env is where the integration stack runs: a DinD box (the default, used on the test server)
// or the machine running go test (OLAKE_TEST_ENV=runner, used on CI runners).
type Env interface {
	// Exec runs cmd with /bin/sh -c and returns its exit code and output.
	Exec(ctx context.Context, cmd string) (int, []byte, error)
	// ExecStream is Exec, logging every output line as it arrives.
	ExecStream(ctx context.Context, t *testing.T, cmd string) (int, string, error)
	// Root is the repository root as Exec sees it.
	Root() string
	// Addr returns the host and port the Go test reaches a published port on, e.g. "5433/tcp".
	Addr(ctx context.Context, port string) (host, hostPort string, err error)
	// InstallPlaywright installs the UI dependencies and Chromium, unless CI already has.
	InstallPlaywright(ctx context.Context, t *testing.T) error
}

// newEnv starts the Env that OLAKE_TEST_ENV selects. ports are the published ports the Go test
// reaches through Addr.
func newEnv(ctx context.Context, t *testing.T, ports []string) (Env, error) {
	switch mode := os.Getenv("OLAKE_TEST_ENV"); mode {
	case "", "dind":
		return startBoxEnv(ctx, t, ports)
	default:
		return nil, fmt.Errorf("unknown OLAKE_TEST_ENV %q: want dind", mode)
	}
}
