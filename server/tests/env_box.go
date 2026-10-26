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

	"github.com/moby/moby/api/types/container"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// boxRoot is where the box mounts the repository.
const boxRoot = "/mnt"

// Install required tools during initial setup
const setupToolsCmd = `
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

// boxEnv runs the stack inside a privileged ubuntu:22.04 container with its own dockerd, with
// the repository mounted at /mnt. It is the default Env.
type boxEnv struct {
	ctr testcontainers.Container
}

// startBoxEnv starts the box, publishing ports, and installs the tools the run needs in it.
// The box is terminated when the test ends.
func startBoxEnv(ctx context.Context, t *testing.T, ports []string) (*boxEnv, error) {
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return nil, fmt.Errorf("could not determine project root: %w", err)
	}
	t.Logf("Project root identified at: %s", projectRoot)

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
				fmt.Sprintf("%s:%s:rw", projectRoot, boxRoot),
			}
			// Tmpfs mounts create temporary in-memory filesystems inside the container.
			// These directories behave like RAM disks they exist only in memory (not on disk) and are automatically cleaned up when the container stops.
			// This is useful for high-performance temporary storage.
			// 70GB for docker and 15GB for shared memory space in Linux
			hc.Tmpfs = map[string]string{
				"/var/lib/docker": "size=70G",
				"/dev/shm":        "size=10G",
			}
			hc.Memory = 22 * 1024 * 1024 * 1024 // 22GB
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		},
		ConfigModifier: func(config *container.Config) {
			config.WorkingDir = boxRoot
		},
		// Only the ports the Go test reaches. The UI (8000) is used from inside the box, and the box's
		// Docker API listens on 127.0.0.1 only: published, it would give anyone who can reach the
		// runner an unauthenticated daemon in a privileged container.
		ExposedPorts: ports,
		Cmd: []string{
			"/bin/sh", "-c",
			`set -e
			apt-get update -y &&
			apt-get install -y --no-install-recommends ca-certificates curl gnupg lsb-release iproute2 procps &&
			apt-get install -y --no-install-recommends docker.io &&
			mkdir -p /var/lib/docker /var/run/docker &&
			exec dockerd --host=tcp://127.0.0.1:2375 --host=unix:///var/run/docker.sock
			`,
		},
		WaitingFor: wait.ForExec([]string{"docker", "-H", "tcp://127.0.0.1:2375", "info"}).WithStartupTimeout(60 * time.Second).WithPollInterval(1 * time.Second),
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	// Ryuk is disabled, so nothing else removes the box. Terminate it when the test ends, even
	// after a failed step. TerminateContainer is nil-safe, and this comes before the error check
	// because GenericContainer can return a started container together with an error. A failed
	// terminate is only logged, as before.
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("Warning: failed to terminate container: %v", err)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start DinD container: %w", err)
	}
	b := &boxEnv{ctr: ctr}

	t.Log("Docker daemon is ready")
	t.Log("Installing required tools...")
	if code, out, err := b.Exec(ctx, setupToolsCmd); err != nil || code != 0 {
		return nil, fmt.Errorf("tools installation failed (%d): %s\n%s", code, err, out)
	}
	return b, nil
}

func (b *boxEnv) Root() string { return boxRoot }

func (b *boxEnv) Addr(ctx context.Context, port string) (string, string, error) {
	host, err := b.ctr.Host(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get host: %w", err)
	}
	mapped, err := b.ctr.MappedPort(ctx, port)
	if err != nil {
		return "", "", fmt.Errorf("failed to get mapped port for %s: %w", port, err)
	}
	return host, mapped.Port(), nil
}

// ServiceHost is the box's own address: the driver runs on the box's dockerd and reaches the
// ports published on the box through it.
func (b *boxEnv) ServiceHost(ctx context.Context) (string, error) {
	return serviceHost(ctx, b.ctr.ContainerIP)
}

func (b *boxEnv) InstallPlaywright(ctx context.Context, _ *testing.T) error {
	if code, out, err := b.Exec(ctx, installPlaywrightCmd(boxRoot)); err != nil || code != 0 {
		return fmt.Errorf("playwright installation failed (%d): %s\n%s", code, err, out)
	}
	return nil
}

func (b *boxEnv) Exec(ctx context.Context, cmd string) (int, []byte, error) {
	code, reader, err := b.ctr.Exec(ctx, []string{"/bin/sh", "-c", cmd})
	if err != nil {
		return code, nil, err
	}
	output, _ := io.ReadAll(reader)
	return code, output, nil
}

func (b *boxEnv) ExecStream(ctx context.Context, t *testing.T, cmd string) (int, string, error) {
	exitCode, reader, err := b.ctr.Exec(ctx, []string{"sh", "-c", cmd})
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
