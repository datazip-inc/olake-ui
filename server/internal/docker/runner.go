package docker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/mod/semver"
)

// NewRunner creates a new Docker runner
func NewRunner(workingDir string) *Runner {
	if err := utils.CreateDirectory(workingDir, DefaultDirPermissions); err != nil {
		logs.Critical("Failed to create working directory %s: %v", workingDir, err)
	}

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logs.Critical("Failed to create Docker client: %v", err)
	}

	return &Runner{
		WorkingDir:   workingDir,
		anonymousID:  telemetry.GetTelemetryUserID(),
		dockerClient: cli,
	}
}

// Close closes the Docker client connection
func (r *Runner) Close() error {
	if r.dockerClient != nil {
		return r.dockerClient.Close()
	}
	return nil
}

// GetDefaultConfigDir returns the default directory for storing config files
func GetDefaultConfigDir() string {
	// Keep using the docker package default to align with getHostOutputDir mapping rules
	return DefaultConfigDir
}

// setupWorkDirectory creates a working directory and returns the full path
func (r *Runner) setupWorkDirectory(subDir string) (string, error) {
	workDir := filepath.Join(r.WorkingDir, subDir)
	if err := utils.CreateDirectory(workDir, DefaultDirPermissions); err != nil {
		return "", fmt.Errorf("failed to create work directory: %s", err)
	}
	return workDir, nil
}

// writeConfigFiles writes multiple configuration files to the specified directory
func (r *Runner) writeConfigFiles(workDir string, configs []FileConfig) error {
	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := utils.WriteFile(filePath, []byte(config.Data), DefaultFilePermissions); err != nil {
			return fmt.Errorf("failed to write %s: %s", config.Name, err)
		}
	}
	return nil
}

// deleteConfigFiles removes only the config files written in the working directory
func (r *Runner) deleteConfigFiles(workDir string, configs []FileConfig) {
	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := os.Remove(filePath); err != nil {
			logs.Warn("Failed to delete config file %s: %s", filePath, err)
		}
	}
}

// GetDockerImageName constructs a Docker image name based on source type and version
func (r *Runner) GetDockerImageName(sourceType, version string) string {
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
}

// FetchSpec runs the spec command and parses the resulting JSON spec
func (r *Runner) FetchSpec(ctx context.Context, destinationType, sourceType, version, workflowID string) (dto.SpecOutput, error) {
	// For spec, no config file is needed; optionally pass destination-type
	var extra []string
	if strings.TrimSpace(destinationType) != "" {
		extra = append(extra, "--destination-type", destinationType)
	}

	output, err := r.ExecuteDockerCommand(ctx, workflowID, "", Spec, sourceType, version, "", extra...)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("docker command failed: %v", err)
	}
	spec, err := utils.ExtractJSON(string(output))
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to parse spec: %s", string(output))
	}
	return dto.SpecOutput{Spec: spec}, nil
}

// TestConnection runs the check command and returns connection status
func (r *Runner) TestConnection(ctx context.Context, flag, sourceType, version, config, workflowID string) (map[string]interface{}, error) {
	workDir, err := r.setupWorkDirectory(WorkflowHash(workflowID))
	if err != nil {
		return nil, err
	}

	configs := []FileConfig{
		{Name: "config.json", Data: config},
		{Name: "user_id.txt", Data: r.anonymousID},
	}

	if err := r.writeConfigFiles(workDir, configs); err != nil {
		return nil, err
	}
	defer r.deleteConfigFiles(workDir, configs)

	configPath := filepath.Join(workDir, "config.json")
	// flag can be "", "config", etc.; pass as-is
	output, err := r.ExecuteDockerCommand(ctx, WorkflowHash(workflowID), flag, Check, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

	logs.Info("check command output: %s\n", string(output))

	logMsg, err := utils.ExtractJSON(string(output))
	if err != nil {
		return nil, err
	}

	connectionStatus, ok := logMsg["connectionStatus"].(map[string]interface{})
	if !ok || connectionStatus == nil {
		return nil, fmt.Errorf("connection status not found")
	}

	status, statusOk := connectionStatus["status"].(string)
	message, _ := connectionStatus["message"].(string) // message is optional
	if !statusOk {
		return nil, fmt.Errorf("connection status not found")
	}

	return map[string]interface{}{
		"message": message,
		"status":  status,
	}, nil
}

// GetCatalog runs the discover command and returns catalog data
func (r *Runner) GetCatalog(ctx context.Context, sourceType, version, config, workflowID, streamsConfig, jobName string) (map[string]interface{}, error) {
	workDir, err := r.setupWorkDirectory(WorkflowHash(workflowID))
	if err != nil {
		return nil, err
	}
	configs := []FileConfig{
		{Name: "config.json", Data: config},
		{Name: "streams.json", Data: streamsConfig},
		{Name: "user_id.txt", Data: r.anonymousID},
	}

	if err := r.writeConfigFiles(workDir, configs); err != nil {
		return nil, err
	}
	defer r.deleteConfigFiles(workDir, configs)

	configPath := filepath.Join(workDir, "config.json")
	catalogPath := filepath.Join(workDir, "streams.json")

	var catalogsArgs []string
	if streamsConfig != "" {
		catalogsArgs = append(catalogsArgs, "--catalog", "/mnt/config/streams.json")
	}
	if jobName != "" && semver.IsValid(version) && semver.Compare(version, "v0.2.0") >= 0 {
		catalogsArgs = append(catalogsArgs, "--destination-database-prefix", jobName)
	}

	if _, err = r.ExecuteDockerCommand(ctx, WorkflowHash(workflowID), "config", Discover, sourceType, version, configPath, catalogsArgs...); err != nil {
		return nil, err
	}

	// Parse the resulting catalog file
	return utils.ParseJSONFile(catalogPath)
}

// RunSync runs the sync command to transfer data from source to destination
func (r *Runner) RunSync(ctx context.Context, jobID int, workflowID string) (map[string]interface{}, error) {
	// Deterministic container name to enable adoption across retries
	containerName := WorkflowHash(workflowID)

	// Setup work dir and configs
	workDir, err := r.setupWorkDirectory(containerName)
	if err != nil {
		logs.Error("workflowID %s: failed to setup work directory: %s", workflowID, err)
		return nil, err
	}
	logs.Info("working directory path %s", workDir)

	// Marker to indicate we have launched once
	launchedMarker := filepath.Join(workDir, "logs")

	// Inspect container state using Docker SDK
	state := r.getContainerState(ctx, containerName, workflowID)

	// 1) If container is running, adopt and wait for completion
	if state.Exists && state.Running {
		logs.Info("workflowID %s: adopting running container %s", workflowID, containerName)
		if err := r.waitContainer(ctx, containerName, workflowID); err != nil {
			logs.Error("workflowID %s: container wait failed: %s", workflowID, err)
			return nil, err
		}
		state = r.getContainerState(ctx, containerName, workflowID)
	}

	// 2) If container exists and exited, treat as finished: cleanup and return status
	if state.Exists && !state.Running && state.ExitCode != nil {
		logs.Info("workflowID %s: container %s exited with code %d", workflowID, containerName, *state.ExitCode)
		if *state.ExitCode == 0 {
			return map[string]interface{}{"status": "completed"}, nil
		}
		return nil, fmt.Errorf("workflowID %s: container %s exit %d", workflowID, containerName, *state.ExitCode)
	}

	// 3) First launch path: only if we never launched and nothing is running
	if _, err := os.Stat(launchedMarker); os.IsNotExist(err) {
		logs.Info("workflowID %s: first launch path, preparing configs", workflowID)
		jobORM := database.NewJobORM()
		job, err := jobORM.GetByID(jobID, false)
		if err != nil {
			logs.Error("workflowID %s: failed to fetch job %d: %s", workflowID, jobID, err)
			return nil, err
		}
		configs := []FileConfig{
			{Name: "config.json", Data: job.SourceID.Config},
			{Name: "streams.json", Data: job.StreamsConfig},
			{Name: "writer.json", Data: job.DestID.Config},
			{Name: "state.json", Data: job.State},
			{Name: "user_id.txt", Data: r.anonymousID},
		}
		if err := r.writeConfigFiles(workDir, configs); err != nil {
			logs.Error("workflowID %s: failed to write config files: %s", workflowID, err)
			return nil, err
		}

		configPath := filepath.Join(workDir, "config.json")
		logs.Info("workflowID %s: executing docker container %s", workflowID, containerName)

		if _, err = r.ExecuteDockerCommand(
			ctx,
			containerName,
			"config",
			Sync,
			job.SourceID.Type,
			job.SourceID.Version,
			configPath,
			"--catalog", "/mnt/config/streams.json",
			"--destination", "/mnt/config/writer.json",
			"--state", "/mnt/config/state.json",
		); err != nil {
			logs.Error("workflowID %s: docker execution failed: %s", workflowID, err)
			return nil, err
		}

		logs.Info("workflowID %s: container %s completed successfully", workflowID, containerName)
		return map[string]interface{}{"status": "completed"}, nil
	}

	// Skip if container is not running, was already launched (logs exist), and no new run is needed.
	logs.Info("workflowID %s: container %s already handled, skipping launch", workflowID, containerName)
	return map[string]interface{}{"status": "skipped"}, nil
}

type ContainerState struct {
	Exists   bool
	Running  bool
	ExitCode *int
}

func (r *Runner) getContainerState(ctx context.Context, name, workflowID string) ContainerState {
	inspect, err := r.dockerClient.ContainerInspect(ctx, name)
	if err != nil || inspect.ContainerJSONBase == nil || inspect.State == nil {
		logs.Warn("workflowID %s: container inspect failed or state missing for %s: %v", workflowID, name, err)
		return ContainerState{Exists: false}
	}
	running := inspect.State.Running
	var ec *int
	if !running && inspect.State.ExitCode != 0 {
		code := inspect.State.ExitCode
		ec = &code
	}
	return ContainerState{Exists: true, Running: running, ExitCode: ec}
}

func (r *Runner) waitContainer(ctx context.Context, name, workflowID string) error {
	statusCh, errCh := r.dockerClient.ContainerWait(ctx, name, "")
	select {
	case err := <-errCh:
		if err != nil {
			logs.Error("workflowID %s: container wait failed for %s: %s", workflowID, name, err)
			return fmt.Errorf("docker wait failed: %s", err)
		}
		return nil
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("workflowID %s: container %s exited with code %d", workflowID, name, status.StatusCode)
		}
		return nil
	}
}

// StopContainer stops a container by name, falling back to kill if needed.
func StopContainer(ctx context.Context, workflowID string) error {
	containerName := WorkflowHash(workflowID)
	logs.Info("workflowID %s: stop request received for container %s", workflowID, containerName)

	if strings.TrimSpace(containerName) == "" {
		logs.Warn("workflowID %s: empty container name", workflowID)
		return fmt.Errorf("empty container name")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to init docker client: %v", err)
	}
	defer cli.Close()

	// Graceful stop with timeout in seconds (use StopOptions{} if relying on engine default)
	t := 5 // seconds
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &t}); err != nil {
		logs.Warn("workflowID %s: docker stop failed for %s: %s", workflowID, containerName, err)
		if kerr := cli.ContainerKill(ctx, containerName, "SIGKILL"); kerr != nil {
			logs.Error("workflowID %s: docker kill failed for %s: %s", workflowID, containerName, kerr)
			return fmt.Errorf("docker kill failed: %s", kerr)
		}
	}

	// Remove container
	if err := cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}); err != nil {
		logs.Warn("workflowID %s: docker rm failed for %s: %s", workflowID, containerName, err)
	} else {
		logs.Info("workflowID %s: container %s removed successfully", workflowID, containerName)
	}
	return nil
}

// PersistJobStateFromFile reads the state JSON file and updates the job state
func (r *Runner) PersistJobStateFromFile(jobID int, workflowID string) error {
	hashWorkflowID := WorkflowHash(workflowID)
	workDir, err := r.setupWorkDirectory(hashWorkflowID)
	if err != nil {
		logs.Error("workflowID %s: failed to setup work directory: %s", workflowID, err)
		return err
	}

	statePath := filepath.Join(workDir, "state.json")
	state, err := utils.ParseJSONFile(statePath)
	if err != nil {
		logs.Error("workflowID %s: failed to parse state file %s: %s", workflowID, statePath, err)
		return err
	}

	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID, false)
	if err != nil {
		logs.Error("workflowID %s: failed to fetch job %d: %s", workflowID, jobID, err)
		return err
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		logs.Error("workflowID %s: failed to marshal state: %s", workflowID, err)
		return err
	}

	job.State = string(stateJSON)
	job.Active = true

	if err := jobORM.Update(job); err != nil {
		logs.Error("workflowID %s: failed to update job %d: %s", workflowID, jobID, err)
		return err
	}

	logs.Info("workflowID %s: job state persisted successfully for jobID %d", workflowID, jobID)
	return nil
}

// WorkflowHash returns a deterministic hash string for a given workflowID
func WorkflowHash(workflowID string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
}
