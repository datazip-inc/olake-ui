package docker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/utils"
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
func (r *Runner) ExecuteDockerCommand(ctx context.Context, flag string, command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := filepath.Dir(configPath)
	if err := utils.CreateDirectory(outputDir, DefaultDirPermissions); err != nil {
		return nil, err
	}

	dockerArgs := r.buildDockerArgs(ctx, flag, command, sourceType, version, configPath, outputDir, additionalArgs...)
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
func (r *Runner) buildDockerArgs(ctx context.Context, flag string, command Command, sourceType, version, configPath, outputDir string, additionalArgs ...string) []string {
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

	dockerArgs := []string{"run", "--rm"}

	dockerArgs = append(dockerArgs,
		"-v", fmt.Sprintf("%s:/mnt/config", hostOutputDir),
		imageName,
		string(command),
		fmt.Sprintf("--%s", flag), fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)),
	)

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
	output, err := r.ExecuteDockerCommand(ctx, flag, Check, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

	logs.Info("check command output: %s\n", string(output))

	logMsg, err := utils.ExtractAndParseLastLogMessage(output)
	if err != nil {
		return nil, err
	}

	if logMsg.ConnectionStatus == nil {
		return nil, fmt.Errorf("connection status not found")
	}

	return map[string]interface{}{
		"message": logMsg.ConnectionStatus.Message,
		"status":  logMsg.ConnectionStatus.Status,
	}, nil
}

// GetCatalog runs the discover command and returns catalog data
func (r *Runner) GetCatalog(ctx context.Context, sourceType, version, config, workflowID, streamsConfig string) (map[string]interface{}, error) {
	workDir, err := r.setupWorkDirectory(workflowID)
	if err != nil {
		return nil, err
	}
	logs.Info("working directory path %s\n", workDir)

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
		catalogsArgs = []string{
			"--catalog", "/mnt/config/streams.json",
		}
	}
	_, err = r.ExecuteDockerCommand(ctx, "config", Discover, sourceType, version, configPath, catalogsArgs...)
	if err != nil {
		return nil, err
	}

	// Simplified JSON parsing - just parse if exists, return error if not
	return utils.ParseJSONFile(catalogPath)
}

// RunSync runs the sync command to transfer data from source to destination
func (r *Runner) RunSync(ctx context.Context, jobID int, workflowID string) (map[string]interface{}, error) {
	// Generate unique directory name
	workDir, err := r.setupWorkDirectory(fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID))))
	if err != nil {
		return nil, err
	}
	logs.Info("working directory path %s\n", workDir)
	// Get current job state
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID, false)
	if err != nil {
		return nil, err
	}

	// Prepare all configuration files
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
	statePath := filepath.Join(workDir, "state.json")

	// Execute sync command
	_, err = r.ExecuteDockerCommand(ctx, "config", Sync, job.SourceID.Type, job.SourceID.Version, configPath,
		"--catalog", "/mnt/config/streams.json",
		"--destination", "/mnt/config/writer.json",
		"--state", "/mnt/config/state.json")
	if err != nil {
		return nil, err
	}
	// Parse state file
	result, err := utils.ParseJSONFile(statePath)
	if err != nil {
		return nil, err
	}

	// Update job state if we have valid result
	if stateJSON, err := json.Marshal(result); err == nil {
		job.State = string(stateJSON)
		job.Active = true
		if err := jobORM.Update(job); err != nil {
			return nil, err
		}
	}
	return result, nil
}
