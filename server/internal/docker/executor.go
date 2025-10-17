package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

// pullImage pulls the Docker image if needed
func (r *Runner) pullImage(ctx context.Context, imageName, version string) error {
	// Only pull if version is "latest"; otherwise assume preloaded or let failure surface
	if version == "latest" {
		logs.Info("Pulling Docker image: %s", imageName)
		reader, err := r.dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image %s: %s", imageName, err)
		}
		defer reader.Close()

		// Read the pull output (optional, for logging)
		if _, err = io.Copy(io.Discard, reader); err != nil {
			logs.Warning("Failed to read image pull output: %s", err)
		}
	}
	return nil
}

// ExecuteDockerCommand executes a Docker command with the given parameters using Docker SDK
// containerName is used for deterministic adoption/stop flows.
func (r *Runner) ExecuteDockerCommand(ctx context.Context, containerName, flag string, command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := "."
	if configPath != "" {
		outputDir = filepath.Dir(configPath)
		if err := utils.CreateDirectory(outputDir, DefaultDirPermissions); err != nil {
			return nil, err
		}
	}

	imageName := r.GetDockerImageName(sourceType, version)

	// Pull image if necessary
	if err := r.pullImage(ctx, imageName, version); err != nil {
		return nil, err
	}

	// Build command arguments
	cmdArgs := r.buildCommandArgs(flag, command, configPath, additionalArgs...)

	// Get host output directory for volume mounting
	hostOutputDir := r.getHostOutputDir(outputDir)

	// Environment variables propagation
	var envs []string
	for k, v := range utils.GetWorkerEnvVars() {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   cmdArgs,
		Env:   envs,
	}

	// Create host configuration with volume mounts
	var mounts []mount.Mount
	if hostOutputDir != "" {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostOutputDir,
			Target: ContainerMountDir,
		})
	}
	hostConfig := &container.HostConfig{
		Mounts:     mounts,
		AutoRemove: true, // Automatically remove container when it exits
	}

	logs.Info("Running Docker container with image: %s, name: %s, command: %v", imageName, containerName, cmdArgs)

	// Create container with deterministic name (may already exist if adopted)
	resp, err := r.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already in use") {
		return nil, fmt.Errorf("failed to create container: %s", err)
	}
	id := resp.ID
	if id == "" {
		// Container might already exist; use the name for subsequent ops
		id = containerName
	}

	// Start container
	if err := r.dockerClient.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		// If it's already running, continue
		if !strings.Contains(strings.ToLower(err.Error()), "already started") {
			return nil, fmt.Errorf("failed to start container: %s", err)
		}
	}

	// Wait for container to finish
	statusCh, errCh := r.dockerClient.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %s", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			// Get container logs for error details
			logOutput, _ := r.getContainerLogs(ctx, id)
			return nil, fmt.Errorf("container exited with status %d: %s", status.StatusCode, logOutput)
		}
	}

	// Get container logs
	output, err := r.getContainerLogs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %s", err)
	}

	logs.Info("Docker container output: %s", string(output))

	return output, nil
}

// getContainerLogs retrieves logs from a container
func (r *Runner) getContainerLogs(ctx context.Context, containerID string) ([]byte, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	reader, err := r.dockerClient.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, reader); err != nil {
		return nil, err
	}
	// Prefer stdout, but include stderr if present
	if stderrBuf.Len() > 0 && stdoutBuf.Len() == 0 {
		return stderrBuf.Bytes(), nil
	}
	if stderrBuf.Len() > 0 {
		return append(stdoutBuf.Bytes(), []byte("\n"+stderrBuf.String())...), nil
	}
	return stdoutBuf.Bytes(), nil
}

// buildCommandArgs constructs the command arguments for the container
func (r *Runner) buildCommandArgs(flag string, command Command, configPath string, additionalArgs ...string) []string {
	cmdArgs := []string{string(command)}

	if strings.TrimSpace(flag) != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--%s", flag))
	}
	if strings.TrimSpace(configPath) != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("%s/%s", ContainerMountDir, filepath.Base(configPath)))
	}

	// Add encryption key as a flag if it exists (preserve original behavior)
	if encryptionKey := os.Getenv(constants.EncryptionKey); encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	return append(cmdArgs, additionalArgs...)
}

// getHostOutputDir determines the host output directory path
func (r *Runner) getHostOutputDir(outputDir string) string {
	if persistentDir := os.Getenv(envPersistentDir); persistentDir != "" {
		hostOutputDir := strings.Replace(outputDir, DefaultConfigDir, persistentDir, 1)
		logs.Info("hostOutputDir %s", hostOutputDir)
		return hostOutputDir
	}
	return outputDir
}
