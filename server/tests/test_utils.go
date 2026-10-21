package tests

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/apache/spark-connect-go/v35/spark/sql"
	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/datazip-inc/olake-ui/server/tests/destinations"
	"github.com/datazip-inc/olake-ui/server/tests/sources"
)

const (
	// Install required tools during initial setup
	setupToolsCmd = `
        apt-get update && 
        apt-get install -y curl postgresql-client dnsutils iputils-ping ncurses-bin ca-certificates gnupg lsb-release netcat-openbsd libatomic1 && 
        curl -fsSL https://deb.nodesource.com/setup_22.x | bash - &&
        apt-get install -y nodejs &&
        npm install -g pnpm@11.24.0 &&
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg &&
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null &&
        apt-get update &&
        apt-get install -y docker-compose-plugin &&
        update-ca-certificates
    `

	// Start OLake application
	startOLakeCmd = `
        cd /mnt && 
        mkdir -p /mnt/olake-data && 
        docker compose up -d && 
        for i in $(seq 1 60); do
            if curl -f http://localhost:8000/health 2>/dev/null || curl -f http://localhost:8000 2>/dev/null; then
                echo "OLake UI ready."
                break
            fi
            sleep 2
        done
    `

	// Install Playwright and dependencies. --frozen-lockfile installs the Playwright version
	// pinned in pnpm-lock.yaml instead of resolving the latest release.
	installPlaywrightCmd = `
        cd /mnt/ui &&
        pnpm install --frozen-lockfile &&
        pnpm exec playwright install --with-deps chromium
    `

	// Run a Playwright spec, relative to ui/
	runPlaywrightCmd = `
        cd /mnt/ui &&
        E2E_JOB_NAME='%s' PLAYWRIGHT_TEST_BASE_URL=http://localhost:8000 DEBUG=pw:api npx playwright test %s
    `
)

// sparkConnectPort is the box port VerifyDestination reaches Spark Connect on. Every
// destination's compose file must run Spark Connect there; see destinations.Destination.
const sparkConnectPort = "15002/tcp"

// Workflow is one integration run: the source and destination to start, the Playwright spec
// that creates and syncs the job through the UI, and the checks run on the destination afterwards.
type Workflow struct {
	Source      sources.Source
	Destination destinations.Destination
	Spec        string  // Playwright spec relative to ui/
	Job         string  // name of the job the spec creates; passed to it as E2E_JOB_NAME
	Checks      []Check // run on the destination after the spec has passed
}

// database is the destination database OLake writes the job to: <job>_<source namespace>, lowercased.
func (w Workflow) database() string {
	return strings.ToLower(w.Job + "_" + w.Source.Namespace)
}

// Check verifies the destination through Spark. table turns a source table name into the
// fully qualified name Spark reads it from.
type Check func(ctx context.Context, t *testing.T, spark sql.SparkSession, table func(name string) string)

// composeStartCmd starts the compose file in dir (relative to server/tests) as project
// olake-<project>-test, runs ready in a subshell (skipped when empty), then lists the services.
// project is "source" or "destination", so the two never share a project whatever their names.
func composeStartCmd(project, dir string, services []string, ready string) string {
	name := fmt.Sprintf("olake-%s-test", project)
	cmd := fmt.Sprintf("cd /mnt/server/tests/%s && docker compose -p %s up -d %s", dir, name, strings.Join(services, " "))
	if strings.TrimSpace(ready) != "" {
		cmd += " && (" + ready + ")"
	}
	return cmd + " && docker compose -p " + name + " ps"
}

func DinDTestContainer(t *testing.T, w Workflow) error {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute) // under the CI job's 60 minutes
	defer cancel()
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return fmt.Errorf("could not determine project root: %w", err)
	}
	t.Logf("Project root identified at: %s", projectRoot)

	exposedPorts := []string{"8000/tcp", "2375/tcp", sparkConnectPort}
	if w.Source.Port != "" {
		exposedPorts = append(exposedPorts, w.Source.Port)
	}

	req := testcontainers.ContainerRequest{
		Image: "ubuntu:22.04",
		Env: map[string]string{
			"DOCKER_TLS_CERTDIR":           "",
			"TELEMETRY_DISABLED":           "true",
			"TESTCONTAINERS_RYUK_DISABLED": "true",
			"DEBIAN_FRONTEND":              "noninteractive",
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.Binds = []string{
				fmt.Sprintf("%s:/mnt:rw", projectRoot),
			}
			// Tmpfs mounts create temporary in-memory filesystems inside the container.
			// These directories behave like RAM disks they exist only in memory (not on disk) and are automatically cleaned up when the container stops.
			// This is useful for high-performance temporary storage.
			// 70GB for docker and 15GB for shared memory space in Linux
			hc.Tmpfs = map[string]string{
				"/var/lib/docker": "size=70G",
				"/dev/shm":        "size=10G",
			}
			hc.Resources.Memory = 22 * 1024 * 1024 * 1024 // 22GB
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		},
		ConfigModifier: func(config *container.Config) {
			config.WorkingDir = "/mnt"
		},
		ExposedPorts: exposedPorts,
		Cmd: []string{
			"/bin/sh", "-c",
			`set -e
			apt-get update -y &&
			apt-get install -y --no-install-recommends ca-certificates curl gnupg lsb-release iproute2 procps &&
			apt-get install -y --no-install-recommends docker.io &&
			mkdir -p /var/lib/docker /var/run/docker &&
			exec dockerd --host=tcp://0.0.0.0:2375 --host=unix:///var/run/docker.sock
			`,
		},
		WaitingFor: wait.ForExec([]string{"docker", "-H", "tcp://127.0.0.1:2375", "info"}).WithStartupTimeout(60 * time.Second).WithPollInterval(1 * time.Second),
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	// Ryuk is disabled, so nothing else removes the box. CleanupContainer terminates it when the
	// test ends, even after a failed step. It is nil-safe and comes before the error check because
	// GenericContainer can return a started container together with an error.
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		return fmt.Errorf("failed to start DinD container: %w", err)
	}

	// container host
	host, err := ctr.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	// source port, if the source publishes one
	var sourcePort string
	if w.Source.Port != "" {
		mapped, err := ctr.MappedPort(ctx, w.Source.Port)
		if err != nil {
			return fmt.Errorf("failed to get %s port: %w", w.Source.Name, err)
		}
		sourcePort = mapped.Port()
	}

	// Spark Connect port
	sparkPort, err := ctr.MappedPort(ctx, sparkConnectPort)
	if err != nil {
		return fmt.Errorf("failed to get spark port: %w", err)
	}

	t.Log("Docker daemon is ready")
	// Step 1: Install tools
	t.Log("Installing required tools...")
	if code, out, err := ExecCommand(ctx, ctr, setupToolsCmd); err != nil || code != 0 {
		return fmt.Errorf("tools installation failed (%d): %s\n%s", code, err, out)
	}

	// Step 2: Start the source
	t.Logf("Starting %s source...", w.Source.Name)
	if code, out, err := ExecCommand(ctx, ctr, composeStartCmd("source", w.Source.ComposeDir, nil, w.Source.ReadyCmd)); err != nil || code != 0 {
		return fmt.Errorf("%s startup failed (%d): %s\n%s", w.Source.Name, code, err, out)
	}

	// Step 3: Start the destination
	t.Logf("Starting %s destination...", w.Destination.Name)
	if code, out, err := ExecCommand(ctx, ctr, composeStartCmd("destination", w.Destination.ComposeDir, w.Destination.Services, w.Destination.ReadyCmd)); err != nil || code != 0 {
		return fmt.Errorf("destination services startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 4: Patch docker-compose for local images
	t.Log("Patching docker-compose to build local images...")
	if err := PatchDockerCompose(ctx, t, ctr); err != nil {
		return err
	}

	// Step 5: Start OLake application
	t.Log("Starting OLake docker-compose services...")
	if code, out, err := ExecCommand(ctx, ctr, startOLakeCmd); err != nil || code != 0 {
		return fmt.Errorf("OLake startup failed (%d): %s\n%s", code, err, out)
	}

	// Step 6: Seed the source
	w.Source.Seed(ctx, t, host, sourcePort)

	t.Logf("OLake UI is ready and accessible at: http://localhost:8000")

	// Step 7: Install Playwright
	t.Log("Installing Playwright and dependencies...")
	if code, out, err := ExecCommand(ctx, ctr, installPlaywrightCmd); err != nil || code != 0 {
		return fmt.Errorf("playwright installation failed (%d): %s\n%s", code, err, out)
	}

	// Step 8: Run Playwright tests
	t.Log("Executing Playwright tests...")
	if code, out, err := ExecCommandWithStreaming(ctx, t, ctr, fmt.Sprintf(runPlaywrightCmd, w.Job, w.Spec)); err != nil || code != 0 {
		return fmt.Errorf("playwright tests failed (%d): %s\n%s", code, err, out)
	}
	t.Log("Playwright tests passed successfully.")

	// wait before verifying destination data
	t.Log("Waiting for 60 seconds before verifying destination data...")
	time.Sleep(60 * time.Second)

	// Step 9: Verify the destination
	t.Logf("Starting %s data verification...", w.Destination.Name)
	VerifyDestination(ctx, t, host, sparkPort.Port(), w)
	return nil
}

// ExecCommandWithStreaming executes a command and streams output in real-time
func ExecCommandWithStreaming(ctx context.Context, t *testing.T, ctr testcontainers.Container, cmd string) (int, string, error) {
	exitCode, reader, err := ctr.Exec(ctx, []string{"sh", "-c", cmd})
	if err != nil {
		return -1, "", err
	}

	var output strings.Builder
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		t.Log(line)
		output.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		return exitCode, output.String(), err
	}

	return exitCode, output.String(), nil
}

// PatchDockerCompose updates olake-ui to build from local code
// TODO: Remove patch command and find alternative to use local code
func PatchDockerCompose(ctx context.Context, t *testing.T, ctr testcontainers.Container) error {
	patchCmd := `
    set -e
    tmpfile=$(mktemp)
    awk '
    BEGIN{svc="";}
    /^  olake-ui:/{svc="olake-ui"; print; next}
    /^  temporal-worker:/{svc="temporal-worker"; print; next}
    /^  [A-Za-z0-9_-]+:/{ if (svc!="") svc=""; print; next}
    {
      if (svc=="olake-ui" && $0 ~ /^    image:/) {
        print "    build:";
        print "      context: .";
        print "      dockerfile: Dockerfile";
        next
      }
      print
    }
    ' /mnt/docker-compose.yml > "$tmpfile" && mv "$tmpfile" /mnt/docker-compose.yml
`

	code, out, err := ExecCommand(ctx, ctr, patchCmd)
	if err != nil || code != 0 {
		t.Logf("docker-compose patch output: %s", string(out))
		return fmt.Errorf("failed to patch docker-compose.yml (%d): %s\n%s", code, err, out)
	}
	t.Log("docker-compose.yml patched to build local images")
	t.Logf("Patched docker-compose.yml:\n%s", string(out))

	return nil
}

// VerifyDestination connects to Spark and runs every check of the workflow.
func VerifyDestination(ctx context.Context, t *testing.T, host, port string, w Workflow) {
	sparkConnectAddress := fmt.Sprintf("sc://%s:%s", host, port)
	spark, err := sql.NewSessionBuilder().Remote(sparkConnectAddress).Build(ctx)
	require.NoError(t, err, "Failed to connect to Spark Connect server")
	defer func() {
		if stopErr := spark.Stop(); stopErr != nil {
			t.Errorf("Failed to stop Spark session: %v", stopErr)
		}
	}()
	table := func(name string) string { return w.Destination.Table(w.database(), name) }
	for _, check := range w.Checks {
		check(ctx, t, spark, table)
	}
}

// Helper function to execute container commands
func ExecCommand(
	ctx context.Context,
	c testcontainers.Container,
	cmd string,
) (int, []byte, error) {
	code, reader, err := c.Exec(ctx, []string{"/bin/sh", "-c", cmd})
	if err != nil {
		return code, nil, err
	}
	output, _ := io.ReadAll(reader)
	return code, output, nil
}
