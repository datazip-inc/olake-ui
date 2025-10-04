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
	containerName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))

	// Setup work dir and configs
	workDir, err := r.setupWorkDirectory(containerName)
	if err != nil {
		return nil, err
	}

	// Always persist state on exit (success/failure/cancel)
	defer func() {
		if err := r.PersistJobStateFromFile(jobID, workflowID); err != nil {
			logs.Error("Failed to persist state in defer: %v", err)
		}
	}()

	// Marker to indicate we have launched once; prevents relaunch after retries
	launchedMarker := filepath.Join(workDir, "logs")

	// Inspect container state
	state, err := getContainerState(ctx, containerName)
	if err != nil {
		return nil, err
	}

	// 1) If container is running, adopt and wait for completion
	if state.Exists && state.Running {
		logs.Info("Adopting running container %s", containerName)
		if err := waitContainer(ctx, containerName); err != nil {
			return nil, temporal.NewNonRetryableApplicationError(
				err.Error(), "ContainerExitNonZero", err,
			)
		}
		// exit code now available; fall through to finished handling below
		state, _ = getContainerState(ctx, containerName)
	}

	// 2) If container exists and exited, treat as finished: cleanup and return status
	if state.Exists && !state.Running && state.ExitCode != nil {
		logs.Info("Container %s exited with code %d", containerName, *state.ExitCode)
		if *state.ExitCode == 0 {
			return map[string]interface{}{"status": "success"}, nil
		}
		// Return typed error so policy can decide retry vs. fail-fast
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("container %s exit %d", containerName, *state.ExitCode),
			"ContainerExitNonZero",
			nil,
		)
	}

	// 3) First launch path: only if we never launched and nothing is running
	if _, err := os.Stat(launchedMarker); os.IsNotExist(err) {
		jobORM := database.NewJobORM()
		job, err := jobORM.GetByID(jobID, false)
		if err != nil {
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
			return nil, err
		}

		configPath := filepath.Join(workDir, "config.json")

		// Important: ExecuteDockerCommand must run with --name containerName internally
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
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ContainerExitNonZero", err)
		}
		return map[string]interface{}{"status": "launched"}, nil
	}

	// Should not reach here, but in case: do not relaunch
	return map[string]interface{}{"status": "noop"}, nil
}

type ContainerState struct {
	Exists   bool
	Running  bool
	ExitCode *int
}

func getContainerState(ctx context.Context, name string) (ContainerState, error) {
	// docker inspect returns fields if exists
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Status}} {{.State.Running}} {{.State.ExitCode}}", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// treat not found as non-existent
		return ContainerState{Exists: false}, nil
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 3 {
		return ContainerState{Exists: false}, nil
	}
	status := parts[0]
	running := parts[1] == "true"
	var ec *int
	if !running && (status == "exited" || status == "dead") {
		if code, convErr := strconv.Atoi(parts[2]); convErr == nil {
			ec = &code
		}
	}
	return ContainerState{Exists: true, Running: running, ExitCode: ec}, nil
}

func waitContainer(ctx context.Context, name string) error {
	// docker wait prints exit code; validate non-zero as error
	cmd := exec.CommandContext(ctx, "docker", "wait", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker wait failed: %w", err)
	}
	if code, convErr := strconv.Atoi(strings.TrimSpace(string(out))); convErr == nil && code != 0 {
		return fmt.Errorf("container %s exited with code %d", name, code)
	}
	return nil
}

// StopContainer stops a container by name, falling back to kill if needed.
func StopContainer(ctx context.Context, workflowID string) error {
	containerName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	logs.Info("workflow cancel request received for %s, stopping container %s\n", workflowID, containerName)

	if strings.TrimSpace(containerName) == "" {
		return fmt.Errorf("empty container name")
	}

	// Attempt graceful stop
	stopCmd := exec.CommandContext(ctx, "docker", "stop", "-t", "5", containerName)
	if out, err := stopCmd.CombinedOutput(); err != nil {
		logs.Warn("docker stop failed for %s: %v, output: %s", containerName, err, string(out))

		// If stop fails, force kill
		killCmd := exec.CommandContext(ctx, "docker", "kill", containerName)
		if kout, kerr := killCmd.CombinedOutput(); kerr != nil {
			logs.Error("docker kill failed for %s: %v, output: %s", containerName, kerr, string(kout))
		}
	}

	// Always attempt cleanup to remove container
	rmCmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerName)
	if rmOut, rmErr := rmCmd.CombinedOutput(); rmErr != nil {
		logs.Warn("docker rm failed for %s: %v, output: %s", containerName, rmErr, string(rmOut))
	} else {
		logs.Info("Container %s removed successfully", containerName)
	}

	return nil
}

// PersistJobStateFromFile reads the state JSON file and updates the job state
func (r *Runner) PersistJobStateFromFile(jobID int, workflowID string) error {
	hashWorkflowID := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	workDir, err := r.setupWorkDirectory(hashWorkflowID)

	statePath := filepath.Join(workDir, "state.json")
	if err != nil {
		return fmt.Errorf("failed to read state file %s: %s", statePath, err)
	}

	state, err := utils.ParseJSONFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file %s: %s", statePath, err)
	}

	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID, false)
	if err != nil {
		return fmt.Errorf("failed to fetch job %d: %s", jobID, err)
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %s", err)
	}

	job.State = string(stateJSON)
	job.Active = true

	if err := jobORM.Update(job); err != nil {
		return fmt.Errorf("failed to update job with id: %d: %s", jobID, err)
	}

	logs.Info("Job state persisted successfully for jobID %d", jobID)
	return nil
}
