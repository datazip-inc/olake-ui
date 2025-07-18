package docker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beego/beego/v2/core/logs"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/utils"
	"github.com/docker/docker/client"
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
	return constants.DefaultConfigDir
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
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("%s-%s:%s", dockerImagePrefix, sourceType, version)
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
	defer r.deleteConfigFiles(workDir, configs)

	configPath := filepath.Join(workDir, "config.json")
	output, err := r.ExecuteDockerCommand(ctx, flag, Check, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

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
	logs.Info("working directory path %s", workDir)

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
		catalogsArgs = []string{
			"--catalog", "/mnt/config/streams.json",
		}
	}

	_, err = r.ExecuteDockerCommand(ctx, "config", Discover, sourceType, version, configPath, catalogsArgs...)
	if err != nil {
		return nil, err
	}

	// Parse the resulting catalog file
	return utils.ParseJSONFile(catalogPath)
}

// RunSync runs the sync command to transfer data from source to destination
func (r *Runner) RunSync(ctx context.Context, jobID int, workflowID string) (map[string]interface{}, error) {
	// Generate unique directory name
	workDir, err := r.setupWorkDirectory(fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID))))
	if err != nil {
		return nil, err
	}
	logs.Info("working directory path %s", workDir)

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
	defer r.deleteConfigFiles(workDir, configs)

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
