package docker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/utils"
	"go.temporal.io/sdk/temporal"
	"golang.org/x/mod/semver"
)

// Constants
const (
	DefaultDirPermissions  = 0755
	DefaultFilePermissions = 0644
)

// Command represents a Docker command type
type Command string

const (
	Discover Command = "discover"
	Spec     Command = "spec"
	Check    Command = "check"
	Sync     Command = "sync"
)

// File configuration for different operations
type FileConfig struct {
	Name string
	Data string
}

// Runner is responsible for executing Docker commands
type Runner struct {
	WorkingDir  string
	anonymousID string
}

// NewRunner creates a new Docker runner
func NewRunner(workingDir string) *Runner {
	if err := utils.CreateDirectory(workingDir, DefaultDirPermissions); err != nil {
		logs.Critical("Failed to create working directory %s: %v", workingDir, err)
	}

	return &Runner{
		WorkingDir:  workingDir,
		anonymousID: telemetry.GetTelemetryUserID(),
	}
}

// GetDefaultConfigDir returns the default directory for storing config files
func GetDefaultConfigDir() string {
	return constants.DefaultConfigDir
}

// setupWorkDirectory creates a working directory and returns the full path
func (r *Runner) setupWorkDirectory(subDir string) (string, error) {
	workDir := filepath.Join(r.WorkingDir, subDir)
	if err := utils.CreateDirectory(workDir, DefaultDirPermissions); err != nil {
		return "", fmt.Errorf("failed to create work directory: %v", err)
	}
	return workDir, nil
}

// writeConfigFiles writes multiple configuration files to the specified directory
func (r *Runner) writeConfigFiles(workDir string, configs []FileConfig) error {
	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := utils.WriteFile(filePath, []byte(config.Data), DefaultFilePermissions); err != nil {
			return fmt.Errorf("failed to write %s: %v", config.Name, err)
		}
	}
	return nil
}

// GetDockerImageName constructs a Docker image name based on source type and version
func (r *Runner) GetDockerImageName(sourceType, version string) string {
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
}

// ExecuteDockerCommand executes a Docker command with the given parameters
func (r *Runner) ExecuteDockerCommand(ctx context.Context, containerName, flag string, command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := filepath.Dir(configPath)
	if err := utils.CreateDirectory(outputDir, DefaultDirPermissions); err != nil {
		return nil, err
	}

	dockerArgs := r.buildDockerArgs(ctx, containerName, flag, command, sourceType, version, configPath, outputDir, additionalArgs...)
	if len(dockerArgs) == 0 {
		return nil, fmt.Errorf("failed to build docker args")
	}

	logs.Info("Running Docker command: docker %s\n", strings.Join(dockerArgs, " "))

	dockerCmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	output, err := dockerCmd.CombinedOutput()

	logs.Info("Docker command output: %s\n", string(output))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("docker command failed with exit status %d", exitErr.ExitCode())
		}
		return nil, err
	}

	return output, nil
}

// buildDockerArgs constructs Docker command arguments
func (r *Runner) buildDockerArgs(ctx context.Context, containerName, flag string, command Command, sourceType, version, configPath, outputDir string, additionalArgs ...string) []string {
	hostOutputDir := r.getHostOutputDir(outputDir)

	repositoryBase, err := web.AppConfig.String("CONTAINER_REGISTRY_BASE")
	if err != nil {
		logs.Critical("failed to get CONTAINER_REGISTRY_BASE: %s", err)
		return nil
	}
	imageName := r.GetDockerImageName(sourceType, version)

	// If using ECR, ensure login before run
	if strings.Contains(repositoryBase, "ecr") {
		imageName = fmt.Sprintf("%s/%s", repositoryBase, imageName)
		accountID, region, _, err := utils.ParseECRDetails(imageName)
		if err != nil {
			logs.Critical("failed to parse ECR details: %s", err)
			return nil
		}
		if err := utils.DockerLoginECR(ctx, region, accountID); err != nil {
			logs.Critical("failed to login to ECR: %s", err)
			return nil
		}
	}

	// base docker args
	dockerArgs := []string{"run", "--name", containerName}

	if hostOutputDir != "" {
		dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/mnt/config", hostOutputDir))
	}

	for key, value := range utils.GetWorkerEnvVars() {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	dockerArgs = append(dockerArgs, imageName, string(command))

	if flag != "" {
		dockerArgs = append(dockerArgs, fmt.Sprintf("--%s", flag))
	}

	if configPath != "" {
		dockerArgs = append(dockerArgs, fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)))
	}

	if encryptionKey := os.Getenv(constants.EncryptionKey); encryptionKey != "" {
		dockerArgs = append(dockerArgs, "--encryption-key", encryptionKey)
	}

	return append(dockerArgs, additionalArgs...)
}

// getHostOutputDir determines the host output directory path
func (r *Runner) getHostOutputDir(outputDir string) string {
	if persistentDir := os.Getenv("PERSISTENT_DIR"); persistentDir != "" {
		hostOutputDir := strings.Replace(outputDir, constants.DefaultConfigDir, persistentDir, 1)
		logs.Info("hostOutputDir %s\n", hostOutputDir)
		return hostOutputDir
	}
	return outputDir
}

func (r *Runner) FetchSpec(ctx context.Context, destinationType, sourceType, version, workflowID string) (models.SpecOutput, error) {
	flag := utils.Ternary(destinationType != "", "destination-type", "").(string)
	dockerArgs := r.buildDockerArgs(ctx, workflowID, flag, Spec, sourceType, version, "", "", destinationType)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	logs.Info("Running Docker command: docker %s\n", strings.Join(dockerArgs, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return models.SpecOutput{}, fmt.Errorf("docker command failed: %v\nOutput: %s", err, string(output))
	}
	spec, err := utils.ExtractJSON(string(output))
	if err != nil {
		return models.SpecOutput{}, fmt.Errorf("failed to parse spec: %s", string(output))
	}
	return models.SpecOutput{Spec: spec}, nil
}

// TestConnection runs the check command and returns connection status
func (r *Runner) TestConnection(ctx context.Context, flag, sourceType, version, config, workflowID string) (map[string]interface{}, error) {
	workDir, err := r.setupWorkDirectory(workflowID)
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

	configPath := filepath.Join(workDir, "config.json")
	output, err := r.ExecuteDockerCommand(ctx, workflowID, flag, Check, sourceType, version, configPath)
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
	workDir, err := r.setupWorkDirectory(workflowID)
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

	configPath := filepath.Join(workDir, "config.json")
	catalogPath := filepath.Join(workDir, "streams.json")
	var catalogsArgs []string
	if streamsConfig != "" {
		catalogsArgs = append(catalogsArgs, "--catalog", "/mnt/config/streams.json")
	}
	if jobName != "" && semver.Compare(version, "v0.2.0") >= 0 {
		catalogsArgs = append(catalogsArgs, "--destination-database-prefix", jobName)
	}
	_, err = r.ExecuteDockerCommand(ctx, workflowID, "config", Discover, sourceType, version, configPath, catalogsArgs...)
	if err != nil {
		return nil, err
	}

	// Simplified JSON parsing - just parse if exists, return error if not
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
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "SetupWorkDirectoryFailed", err)
	}

	// Marker to indicate we have launched once; prevents relaunch after retries
	launchedMarker := filepath.Join(workDir, "logs")

	// Inspect container state
	state := getContainerState(ctx, containerName, workflowID)

	// 1) If container is running, adopt and wait for completion
	if state.Exists && state.Running {
		logs.Info("workflowID %s: adopting running container %s", workflowID, containerName)
		if err := waitContainer(ctx, containerName, workflowID); err != nil {
			logs.Error("workflowID %s: container wait failed: %s", workflowID, err)
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ContainerExitNonZero", err)
		}
		state = getContainerState(ctx, containerName, workflowID)
	}

	// 2) If container exists and exited, treat as finished: cleanup and return status
	if state.Exists && !state.Running && state.ExitCode != nil {
		logs.Info("workflowID %s: container %s exited with code %d", workflowID, containerName, *state.ExitCode)
		if *state.ExitCode == 0 {
			return map[string]interface{}{"status": "completed"}, nil
		}
		// Return typed error so policy can decide retry vs. fail-fast
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("workflowID %s: container %s exit %d", workflowID, containerName, *state.ExitCode),
			"ContainerExitNonZero",
			nil,
		)
	}

	// 3) First launch path: only if we never launched and nothing is running
	if _, err := os.Stat(launchedMarker); os.IsNotExist(err) {
		logs.Info("workflowID %s: first launch path, preparing configs", workflowID)
		jobORM := database.NewJobORM()
		job, err := jobORM.GetByID(jobID, false)
		if err != nil {
			logs.Error("workflowID %s: failed to fetch job %d: %s", workflowID, jobID, err)
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "FetchJobFailed", err)
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
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "WriteConfigFilesFailed", err)
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
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ContainerExitNonZero", err)
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

func getContainerState(ctx context.Context, name, workflowID string) ContainerState {
	// docker inspect returns fields if exists
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Status}} {{.State.Running}} {{.State.ExitCode}}", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// treat not found as non-existent
		logs.Warn("workflowID %s: docker inspect failed for %s: %s, output: %s", workflowID, name, err, string(out))
		return ContainerState{Exists: false}
	}
	// Split Docker inspect output into fields: status, running flag, and exit code
	// Example: "exited false 137" → parts[0]="exited", parts[1]="false", parts[2]="137"
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 3 {
		return ContainerState{Exists: false}
	}
	// Docker .State.Status can be "created", "running", "paused", "restarting", "removing", "exited", or "dead"; we only handle running vs exited/dead.
	status := parts[0]
	running := parts[1] == "true"
	var ec *int
	if !running && (status == "exited" || status == "dead") {
		if code, convErr := strconv.Atoi(parts[2]); convErr == nil {
			ec = &code
		}
	}
	return ContainerState{Exists: true, Running: running, ExitCode: ec}
}

func waitContainer(ctx context.Context, name, workflowID string) error {
	// docker wait prints exit code; validate non-zero as error
	cmd := exec.CommandContext(ctx, "docker", "wait", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logs.Error("workflowID %s: docker wait failed for %s: %s, output: %s", workflowID, name, err, string(out))
		return fmt.Errorf("docker wait failed: %s", err)
	}
	strOut := strings.TrimSpace(string(out))
	code, convErr := strconv.Atoi(strOut)
	if convErr != nil {
		logs.Error("workflowID %s: failed to parse exit code from docker wait output %q: %s", workflowID, strOut, convErr)
		return fmt.Errorf("failed to parse exit code: %s", convErr)
	}

	if code != 0 {
		return fmt.Errorf("workflowID %s: container %s exited with code %d", workflowID, name, code)
	}
	return nil
}

// StopContainer stops a container by name, falling back to kill if needed.
func StopContainer(ctx context.Context, workflowID string) error {
	containerName := WorkflowHash(workflowID)
	logs.Info("workflowID %s: stop request received for container %s", workflowID, containerName)

	if strings.TrimSpace(containerName) == "" {
		logs.Warn("workflowID %s: empty container name", workflowID)
		return fmt.Errorf("empty container name")
	}

	stopCmd := exec.CommandContext(ctx, "docker", "stop", "-t", "5", containerName)
	if out, err := stopCmd.CombinedOutput(); err != nil {
		logs.Warn("workflowID %s: docker stop failed for %s: %s, output: %s", workflowID, containerName, err, string(out))
		killCmd := exec.CommandContext(ctx, "docker", "kill", containerName)
		if kout, kerr := killCmd.CombinedOutput(); kerr != nil {
			logs.Error("workflowID %s: docker kill failed for %s: %s, output: %s", workflowID, containerName, kerr, string(kout))
		}
	}

	rmCmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerName)
	if rmOut, rmErr := rmCmd.CombinedOutput(); rmErr != nil {
		logs.Warn("workflowID %s: docker rm failed for %s: %s, output: %s", workflowID, containerName, rmErr, string(rmOut))
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
