package docker

import (
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
)

// pullImage pulls the Docker image if needed
func (r *Runner) pullImage(ctx context.Context, imageName string, version string) error {
	// Only pull if version is "latest" or image doesn't exist locally
	if version == "latest" {
		logs.Info("Pulling Docker image: %s", imageName)

		reader, err := r.dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image %s: %v", imageName, err)
		}
		defer reader.Close()

		// Read the pull output (optional, for logging)
		_, err = io.Copy(io.Discard, reader)
		if err != nil {
			logs.Warning("Failed to read image pull output: %v", err)
		}
	}
	return nil
}

// ExecuteDockerCommand executes a Docker command with the given parameters using Docker SDK
func (r *Runner) ExecuteDockerCommand(ctx context.Context, flag string, command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := filepath.Dir(configPath)
	if err := utils.CreateDirectory(outputDir, DefaultDirPermissions); err != nil {
		return nil, err
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

	// Create container configuration
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   cmdArgs,
	}

	// Create host configuration with volume mounts
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: hostOutputDir,
				Target: ContainerMountDir,
			},
		},
		AutoRemove: true, // Automatically remove container when it exits
	}

	logs.Info("Running Docker container with image: %s, command: %v", imageName, cmdArgs)

	// Create container
	resp, err := r.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := r.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for container to finish
	statusCh, errCh := r.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %v", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			// Get container logs for error details
			logs, _ := r.getContainerLogs(ctx, resp.ID)
			return nil, fmt.Errorf("container exited with status %d: %s", status.StatusCode, logs)
		}
	}

	// Get container logs
	output, err := r.getContainerLogs(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %v", err)
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

	return io.ReadAll(reader)
}

// buildCommandArgs constructs the command arguments for the container
func (r *Runner) buildCommandArgs(flag string, command Command, configPath string, additionalArgs ...string) []string {
	cmdArgs := []string{
		string(command),
		fmt.Sprintf("--%s", flag),
		fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)),
	}

	// Add encryption key as a flag if it exists
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
