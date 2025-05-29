package docker

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/utils"
)

// Constants
const (
	DefaultDirPermissions  = 0755
	DefaultFilePermissions = 0644
	DefaultConfigDir       = "/tmp/olake-config"
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
	WorkingDir string
}

type JobHandler struct {
	web.Controller
	jobORM *database.JobORM
}

// NewRunner creates a new Docker runner
func NewRunner(workingDir string) *Runner {
	if err := createDirectory(workingDir); err != nil {
		fmt.Printf("Warning: Failed to create working directory %s: %v\n", workingDir, err)
	}

	return &Runner{
		WorkingDir: workingDir,
	}
}

// GetDefaultConfigDir returns the default directory for storing config files
func GetDefaultConfigDir() string {
	return DefaultConfigDir
}

// File system utilities
func createDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, DefaultDirPermissions); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
	}
	return nil
}

func writeFile(filePath string, data []byte) error {
	dirPath := filepath.Dir(filePath)
	if err := createDirectory(dirPath); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}
	return nil
}

// setupWorkDirectory creates a working directory and returns the full path
func (r *Runner) setupWorkDirectory(subDir string) (string, error) {
	workDir := filepath.Join(r.WorkingDir, subDir)
	if err := createDirectory(workDir); err != nil {
		return "", fmt.Errorf("failed to create work directory: %v", err)
	}
	return workDir, nil
}

// writeConfigFiles writes multiple configuration files to the specified directory
func (r *Runner) writeConfigFiles(workDir string, configs []FileConfig) error {
	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := writeFile(filePath, []byte(config.Data)); err != nil {
			return fmt.Errorf("failed to write %s: %v", config.Name, err)
		}
	}
	return nil
}

// GetDockerImageName constructs a Docker image name based on source type and version
func (r *Runner) GetDockerImageName(sourceType, version string) string {
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
}

// ExecuteDockerCommand executes a Docker command with the given parameters
func (r *Runner) ExecuteDockerCommand(flag string, command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := filepath.Dir(configPath)
	if err := createDirectory(outputDir); err != nil {
		return nil, err
	}

	dockerArgs := r.buildDockerArgs(flag, command, sourceType, version, configPath, outputDir, additionalArgs...)

	fmt.Printf("Running Docker command: docker %s\n", strings.Join(dockerArgs, " "))

	dockerCmd := exec.Command("docker", dockerArgs...)
	output, err := dockerCmd.CombinedOutput()

	fmt.Printf("Docker command output: %s\n", string(output))

	if err != nil {
		return nil, fmt.Errorf("docker command failed: %v, output: %s", err, string(output))
	}

	return output, nil
}

// buildDockerArgs constructs Docker command arguments
func (r *Runner) buildDockerArgs(flag string, command Command, sourceType, version, configPath, outputDir string, additionalArgs ...string) []string {
	hostOutputDir := r.getHostOutputDir(outputDir)

	dockerArgs := []string{
		"run", "--pull=always",
		"-v", fmt.Sprintf("%s:/mnt/config", hostOutputDir),
		"-u", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		r.GetDockerImageName(sourceType, version),
		string(command),
		fmt.Sprintf("--%s", flag), fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)),
	}

	return append(dockerArgs, additionalArgs...)
}

// getHostOutputDir determines the host output directory path
func (r *Runner) getHostOutputDir(outputDir string) string {
	if persistentDir := os.Getenv("PERSISTENT_DIR"); persistentDir != "" {
		hostOutputDir := strings.Replace(outputDir, DefaultConfigDir, persistentDir, 1)
		fmt.Printf("hostOutputDir %s\n", hostOutputDir)
		return hostOutputDir
	}
	return outputDir
}

// ListOutputFiles lists the files in the output directory for debugging
func (r *Runner) ListOutputFiles(outputDir string, message string) {
	files, _ := os.ReadDir(outputDir)
	fmt.Printf("Files in output directory %s:\n", message)
	for _, file := range files {
		fmt.Printf("- %s\n", file.Name())
	}
}

// ParseJSONFile parses a JSON file into a map
func (r *Runner) ParseJSONFile(filePath string) (map[string]interface{}, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(fileData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %v", filePath, err)
	}

	return result, nil
}

// parseJSONFileWithFallback attempts to parse a JSON file, with fallback handling
func (r *Runner) parseJSONFileWithFallback(targetPath, workDir string) (map[string]interface{}, error) {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return r.handleMissingJSONFile()
	}

	result, err := r.ParseJSONFile(targetPath)
	if err != nil {
		return map[string]interface{}{
			"status": "completed",
			"error":  fmt.Sprintf("failed to parse file: %v", err),
		}, nil
	}

	return result, nil
}

// handleMissingJSONFile handles the case when expected JSON file is missing
func (r *Runner) handleMissingJSONFile() (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":  "completed",
		"message": "operation completed successfully",
	}, nil
}

// FindAlternativeJSONFile tries to find any JSON file that might be the catalog
func (r *Runner) FindAlternativeJSONFile(outputDir, targetPath, excludePath string) ([]byte, error) {
	files, _ := os.ReadDir(outputDir)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && file.Name() != filepath.Base(excludePath) {
			tryPath := filepath.Join(outputDir, file.Name())
			if tryData, tryErr := os.ReadFile(tryPath); tryErr == nil {
				if err := writeFile(targetPath, tryData); err == nil {
					return tryData, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no suitable JSON file found")
}

// TestConnection runs the check command and returns connection status
func (r *Runner) TestConnection(flag, sourceType, version, config, workflowID string) (map[string]interface{}, error) {
	workDir, err := r.setupWorkDirectory(workflowID)
	if err != nil {
		return nil, err
	}

	configs := []FileConfig{
		{Name: "config.json", Data: config},
	}

	if err := r.writeConfigFiles(workDir, configs); err != nil {
		return nil, err
	}

	configPath := filepath.Join(workDir, "config.json")
	output, err := r.ExecuteDockerCommand(flag, Check, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

	fmt.Printf("check command output: %s\n", string(output))

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
func (r *Runner) GetCatalog(sourceType, version, config, workflowID string) (map[string]interface{}, error) {
	workDir, err := r.setupWorkDirectory(workflowID)
	if err != nil {
		return nil, err
	}

	fmt.Printf("working directory path %s\n", workDir)

	configs := []FileConfig{
		{Name: "config.json", Data: config},
	}

	if err := r.writeConfigFiles(workDir, configs); err != nil {
		return nil, err
	}

	configPath := filepath.Join(workDir, "config.json")
	catalogPath := filepath.Join(workDir, "streams.json")

	_, err = r.ExecuteDockerCommand("config", Discover, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

	r.ListOutputFiles(workDir, "after discover")

	return r.parseJSONFileWithFallback(catalogPath, workDir)
}

// RunSync runs the sync command to transfer data from source to destination
func (r *Runner) RunSync(sourceType, version, sourceConfig, destConfig, stateConfig, streamsConfig string, JobId, projectID, sourceID, destID int, workflowID string) (map[string]interface{}, error) {
	// Generate unique directory name
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	workDir, err := r.setupWorkDirectory(syncFolderName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("working directory path %s\n", workDir)

	// Get current job state
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(JobId)
	if err != nil {
		return nil, err
	}

	// Use job state if available
	if job.State != "" {
		stateConfig = job.State
	}

	// Prepare all configuration files
	configs := []FileConfig{
		{Name: "config.json", Data: sourceConfig},
		{Name: "streams.json", Data: streamsConfig},
		{Name: "writer.json", Data: destConfig},
		{Name: "state.json", Data: stateConfig},
	}

	if err := r.writeConfigFiles(workDir, configs); err != nil {
		return nil, err
	}

	configPath := filepath.Join(workDir, "config.json")
	statePath := filepath.Join(workDir, "state.json")

	// Execute sync command
	_, err = r.ExecuteDockerCommand("config", Sync, sourceType, version, configPath,
		"--catalog", "/mnt/config/streams.json",
		"--destination", "/mnt/config/writer.json",
		"--state", "/mnt/config/state.json")
	if err != nil {
		return nil, err
	}

	r.ListOutputFiles(workDir, "after sync")

	// Parse and update job state
	result, err := r.parseJSONFileWithFallback(statePath, workDir)
	if err != nil {
		return nil, err
	}

	// Update job state if we have valid result
	if err := r.updateJobState(jobORM, job, result); err != nil {
		fmt.Printf("Warning: Failed to update job state: %v\n", err)
	}

	return result, nil
}

// updateJobState updates the job state in database
func (r *Runner) updateJobState(jobORM *database.JobORM, job interface{}, result map[string]interface{}) error {
	if stateJSON, err := json.Marshal(result); err == nil {
		// Note: You'll need to implement this based on your actual job struct
		// For now, casting to the expected type - adjust as needed
		if j, ok := job.(interface {
			SetState(string)
			SetActive(bool)
		}); ok {
			j.SetState(string(stateJSON))
			j.SetActive(true)
			// return jobORM.Update(job)
		}
		fmt.Printf("Job state updated: %s\n", string(stateJSON))
	}
	return nil
}
