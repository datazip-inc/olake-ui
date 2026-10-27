package tests

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Env runs the integration stack on the machine running go test, with its Docker daemon. It
// expects the caller (the CI workflow) to have installed Node, pnpm, the UI dependencies,
// Playwright and Chromium. Use it only on a machine that is thrown away after the run, such as a
// GitHub-hosted runner: the stack publishes fixed ports and writes into the checkout.
type Env struct {
	root string
}

// newEnv uses the checkout that contains this package. The compose projects it starts are taken
// down when the test ends.
//
// The stack publishes its ports on every interface (Postgres with trust auth, MinIO, Spark, a
// Jupyter server without a token), so it refuses to run anywhere but a GitHub-hosted runner, which
// GitHub sets RUNNER_ENVIRONMENT=github-hosted on. OLAKE_TEST_RUNNER_ANY_HOST=true overrides that
// for a machine you know is isolated.
func newEnv(t *testing.T) (*Env, error) {
	if os.Getenv("RUNNER_ENVIRONMENT") != "github-hosted" && os.Getenv("OLAKE_TEST_RUNNER_ANY_HOST") != "true" {
		return nil, fmt.Errorf("the integration test publishes the test stack's ports on all interfaces and " +
			"runs only on a GitHub-hosted runner (RUNNER_ENVIRONMENT=github-hosted); " +
			"set OLAKE_TEST_RUNNER_ANY_HOST=true to run it on an isolated machine")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return nil, fmt.Errorf("could not determine project root: %w", err)
	}
	t.Logf("Project root identified at: %s", root)
	e := &Env{root: root}
	t.Cleanup(func() { e.down(t) })
	return e, nil
}

// down removes the containers, networks and volumes of the compose projects the run starts.
// Failures are only logged.
func (e *Env) down(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	t.Log("Taking down the compose projects...")
	for _, cmd := range []string{
		fmt.Sprintf("cd %s && %s down -v -t 5 --remove-orphans", shellQuote(e.root), olakeCompose(e.root)),
		"docker compose -p olake-source-test down -v -t 5",
		"docker compose -p olake-destination-test down -v -t 5",
		"docker rm -f " + webhookContainer,
	} {
		if code, out, err := e.Exec(ctx, cmd); err != nil || code != 0 {
			t.Logf("Warning: %q failed (%d): %v\n%s", cmd, code, err, out)
		}
	}
}

// Root is the repository root.
func (e *Env) Root() string { return e.root }

// Addr returns the host and port the Go test reaches a published port on, e.g. "5433/tcp": the
// stack publishes its ports on this machine.
func (e *Env) Addr(port string) (host, hostPort string) {
	number, _, _ := strings.Cut(port, "/")
	return "localhost", number
}

// serviceHostPattern accepts IPv4 addresses and host names. IPv6 is refused: the specs put the
// host into URLs such as http://<host>:9000 without brackets.
var serviceHostPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*$`)

// ServiceHost is the address the sync driver uses to reach the ports the source and the
// destination publish, passed to the Playwright spec as E2E_SERVICE_HOST. It is E2E_SERVICE_HOST
// when that is set, otherwise the gateway of Docker's default bridge, the address of this machine
// as the driver containers see it.
func (e *Env) ServiceHost(ctx context.Context) (string, error) {
	host := strings.TrimSpace(os.Getenv("E2E_SERVICE_HOST"))
	if host == "" {
		code, out, err := e.Exec(ctx, `docker network inspect bridge --format '{{(index .IPAM.Config 0).Gateway}}'`)
		if err != nil || code != 0 {
			return "", fmt.Errorf("failed to detect the service host: docker network inspect bridge failed (%d): %v\n%s", code, err, out)
		}
		host = strings.TrimSpace(string(out))
		if host == "" {
			return "", fmt.Errorf("no service host detected; set E2E_SERVICE_HOST")
		}
	}
	if !serviceHostPattern.MatchString(host) {
		return "", fmt.Errorf("invalid service host %q", host)
	}
	return host, nil
}

// CheckPlaywright checks that the CI workflow installed Playwright in ui/.
func (e *Env) CheckPlaywright(ctx context.Context, t *testing.T) error {
	cmd := fmt.Sprintf("cd %s && pnpm exec playwright --version", shellQuote(e.root+"/ui"))
	if code, out, err := e.Exec(ctx, cmd); err != nil || code != 0 {
		return fmt.Errorf("playwright is not installed in ui/ (%d): %v\n%s\n"+
			"run: pnpm install --frozen-lockfile && pnpm exec playwright install --with-deps chromium", code, err, out)
	}
	t.Log("Using the Playwright installed by the workflow.")
	return nil
}

// shellCommand runs cmd with sh -c in its own process group. When ctx is done the whole group is
// killed: killing only sh would leave children such as docker compose or Playwright running and
// holding the output pipe open, so the call would not return at the deadline.
func shellCommand(ctx context.Context, shell, cmd string) *exec.Cmd {
	c := exec.CommandContext(ctx, shell, "-c", cmd)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGKILL) }
	c.WaitDelay = 10 * time.Second // a process that left the group cannot hold the pipe forever
	return c
}

// Exec runs cmd with /bin/sh -c and returns its exit code and combined output. A non-zero exit is
// reported through the code, not the error.
func (e *Env) Exec(ctx context.Context, cmd string) (int, []byte, error) {
	out, err := shellCommand(ctx, "/bin/sh", cmd).CombinedOutput()
	return exitResult(ctx, err, out)
}

// ExecStream is Exec, logging every output line as it arrives.
func (e *Env) ExecStream(ctx context.Context, t *testing.T, cmd string) (int, string, error) {
	c := shellCommand(ctx, "sh", cmd)
	stdout, err := c.StdoutPipe()
	if err != nil {
		return -1, "", err
	}
	c.Stderr = c.Stdout
	if err := c.Start(); err != nil {
		return -1, "", err
	}
	// Stop reading at the deadline even if a process outside the group still holds the pipe.
	stop := context.AfterFunc(ctx, func() { _ = stdout.Close() })
	defer stop()

	var output strings.Builder
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // Playwright debug lines can be long
	for scanner.Scan() {
		line := scanner.Text()
		t.Log(line)
		output.WriteString(line + "\n")
	}
	scanErr := scanner.Err()
	_, _ = io.Copy(io.Discard, stdout) // keep the child from blocking on a full pipe

	code, _, err := exitResult(ctx, c.Wait(), nil)
	if err == nil {
		err = scanErr
	}
	return code, output.String(), err
}

// exitResult turns the error of a finished command into Exec's (code, output, error).
func exitResult(ctx context.Context, err error, out []byte) (int, []byte, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return -1, out, ctxErr
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), out, nil
	}
	if err != nil {
		return -1, out, err
	}
	return 0, out, nil
}
