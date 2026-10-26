package tests

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
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
	// ServiceHost is the address the sync driver uses to reach the ports the source and the
	// destination publish. It is passed to the Playwright spec as E2E_SERVICE_HOST.
	ServiceHost(ctx context.Context) (string, error)
	// InstallPlaywright installs the UI dependencies and Chromium, unless CI already has.
	InstallPlaywright(ctx context.Context, t *testing.T) error
}

// serviceHostPattern accepts IPv4 and IPv6 addresses and host names, so the value is safe to put
// in a shell command.
var serviceHostPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.:-]*$`)

// serviceHost returns E2E_SERVICE_HOST when it is set, otherwise the address detect finds.
func serviceHost(ctx context.Context, detect func(context.Context) (string, error)) (string, error) {
	host := strings.TrimSpace(os.Getenv("E2E_SERVICE_HOST"))
	if host == "" {
		detected, err := detect(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to detect the service host: %w", err)
		}
		host = strings.TrimSpace(detected)
	}
	if !serviceHostPattern.MatchString(host) {
		return "", fmt.Errorf("invalid service host %q", host)
	}
	return host, nil
}

// newEnv starts the Env that OLAKE_TEST_ENV selects. ports are the published ports the Go test
// reaches through Addr.
func newEnv(ctx context.Context, t *testing.T, ports []string) (Env, error) {
	switch mode := os.Getenv("OLAKE_TEST_ENV"); mode {
	case "", "dind":
		return startBoxEnv(ctx, t, ports)
	case "runner":
		return newRunnerEnv(t)
	default:
		return nil, fmt.Errorf("unknown OLAKE_TEST_ENV %q: want dind or runner", mode)
	}
}
