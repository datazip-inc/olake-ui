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
	// Create the working directory with proper permissions if it doesn't exist
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		// Create with generous permissions (0755) for directories
		err = os.MkdirAll(workingDir, 0755)
		if err != nil {
			fmt.Printf("Warning: Failed to create working directory %s: %v\n", workingDir, err)
		}
	}

	return &Runner{
		WorkingDir: workingDir,
	}
}

// GetDefaultConfigDir returns the default directory for storing config files
func GetDefaultConfigDir() string {
	return "/tmp/olake-config"
}

// Command represents a Docker command type
type Command string

const (
	// Discover command for discovering source schemas
	Discover Command = "discover"
	// Spec command for getting connector specs
	Spec Command = "spec"
	// Check command for testing connection
	Check Command = "check"
	// Sync command for syncing data between source and destination
	Sync Command = "sync"
)

// createDirectory creates a directory with proper permissions and handles errors
func createDirectory(dirPath string) error {
	// Check if directory already exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Create with generous permissions (0755) for directories
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
	}
	return nil
}

// writeFile writes data to a file with proper permissions and handles errors
func writeFile(filePath string, data []byte) error {
	// Ensure directory exists
	dirPath := filepath.Dir(filePath)
	if err := createDirectory(dirPath); err != nil {
		return err
	}

	// Write file with generous permissions (0644) for files
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}
	return nil
}

// WriteToFile writes data to a file with a specific name based on ID
func (r *Runner) WriteToFile(fileData string, ID any) (string, error) {
	// Create directory if not exists
	if err := createDirectory(r.WorkingDir); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	var fileName string
	switch {
	case ID == nil:
		fileName = "writer.json"
	case ID == "streams":
		fileName = "streams.json"
	default:
		fileName = fmt.Sprintf("config-%v.json", ID)
	}

	// Create a file path
	configPath := filepath.Join(r.WorkingDir, fileName)

	// Write config to file with proper permissions
	if err := writeFile(configPath, []byte(fileData)); err != nil {
		return "", err
	}

	return configPath, nil
}

// GetDockerImageName constructs a Docker image name based on source type and version
func (r *Runner) GetDockerImageName(sourceType, version string) string {
	if version == "" {
		return fmt.Sprintf("olakego/source-%s:latest", sourceType)
	}
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
}

// ExecuteDockerCommand executes a Docker command with the given parameters
func (r *Runner) ExecuteDockerCommand(flag string, command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := filepath.Dir(configPath)

	// Ensure output directory exists with proper permissions
	if err := createDirectory(outputDir); err != nil {
		return nil, err
	}

	// Construct Docker command arguments
	var hostOutputDir string
	if os.Getenv("PERSISTENT_DIR") != "" {
		hostOutputDir = strings.Replace(outputDir, "/tmp/olake-config", os.Getenv("PERSISTENT_DIR"), 1)
		fmt.Printf("hostOutputDir %s\n", hostOutputDir)
	} else {
		hostOutputDir = outputDir
	}
	dockerArgs := []string{
		"run", "--pull=always",
		"-v", fmt.Sprintf("%s:/mnt/config", hostOutputDir),
		// Add user mapping to help with permissions in Docker
		"-u", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		r.GetDockerImageName(sourceType, version),
		string(command),
		fmt.Sprintf("--%s", flag), fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)),
	}

	// Add any additional arguments
	dockerArgs = append(dockerArgs, additionalArgs...)

	// Print the Docker command for debugging
	fmt.Printf("Running Docker command: docker %s\n", dockerArgs)

	// Execute Docker command
	dockerCmd := exec.Command("docker", dockerArgs...)

	// Execute Docker command and capture output
	output, err := dockerCmd.CombinedOutput()
	fmt.Printf("Docker command output: %s\n", string(output))

	if err != nil {
		return nil, fmt.Errorf("docker command failed: %v, output: %s", err, string(output))
	}

	return output, nil
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

// FindAlternativeJSONFile tries to find any JSON file that might be the catalog
func (r *Runner) FindAlternativeJSONFile(outputDir, targetPath, excludePath string) ([]byte, error) {
	files, _ := os.ReadDir(outputDir)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" &&
			file.Name() != filepath.Base(excludePath) {
			tryPath := filepath.Join(outputDir, file.Name())
			tryData, tryErr := os.ReadFile(tryPath)
			if tryErr == nil {
				// Copy this file to target path
				err := writeFile(targetPath, tryData)
				if err == nil {
					return tryData, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no suitable JSON file found")
}

// TestSourceConnection run the check command and return connection status
func (r *Runner) TestConnection(flag, sourceType, version, config, workflowID string) (map[string]interface{}, error) {
	// Create directory for output with proper permissions
	checkFolderName := workflowID
	checkDir := filepath.Join(r.WorkingDir, checkFolderName)

	if err := createDirectory(checkDir); err != nil {
		return nil, err
	}
	// Write config to file with proper permissions
	configPath := filepath.Join(checkDir, "config.json")
	if err := writeFile(configPath, []byte(config)); err != nil {
		return nil, err
	}
	// Execute discover command
	output, err := r.ExecuteDockerCommand(flag, Check, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("check command output: %s\n", string(output))
	logMsg, err := utils.ExtractAndParseLastLogMessage([]byte(output))
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
func (r *Runner) GetCatalog(sourceType, version, config string, workflowID string) (map[string]interface{}, error) {
	discoverFolderName := workflowID
	// Create directory for output with proper permissions
	discoverDir := filepath.Join(r.WorkingDir, discoverFolderName)
	fmt.Printf("working directory path %s\n", discoverDir)

	// Ensure the directory exists with proper permissions
	if err := createDirectory(discoverDir); err != nil {
		return nil, err
	}

	// Write config to file with proper permissions
	configPath := filepath.Join(discoverDir, "config.json")
	if err := writeFile(configPath, []byte(config)); err != nil {
		return nil, err
	}

	// Define streams output path
	catalogPath := filepath.Join(discoverDir, "streams.json")

	// Execute discover command
	_, err := r.ExecuteDockerCommand("config", Discover, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

	// List files in output directory for debugging
	r.ListOutputFiles(discoverDir, "after discover")

	// Check if streams file exists
	if _, err := os.Stat(catalogPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "completed",
			"message": "discover completed successfully",
		}, nil
	}

	// Parse streams file
	result, err := r.ParseJSONFile(catalogPath)
	if err != nil {
		return map[string]interface{}{
			"status": "completed",
			"error":  fmt.Sprintf("failed to parse streams file: %v", err),
		}, nil
	}

	return result, nil
}

// RunSync runs the sync command to transfer data from source to destination
func (r *Runner) RunSync(sourceType, version, sourceConfig, destConfig, stateConfig, streamsConfig string, JobId, projectID, sourceID, destID int, workflowID string) (map[string]interface{}, error) {
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
	// Create directory for output with proper permissions
	syncDir := filepath.Join(r.WorkingDir, syncFolderName)
	fmt.Printf("working directory path %s\n", syncDir)

	// Ensure the directory exists with proper permissions
	if err := createDirectory(syncDir); err != nil {
		return nil, err
	}
	var jobORM *database.JobORM
	jobORM = database.NewJobORM()
	job, err := jobORM.GetByID(JobId)
	if err != nil {
		return nil, err
	}
	stateConfig = job.State

	// Define paths for required files
	configPath := filepath.Join(syncDir, "config.json")
	catalogPath := filepath.Join(syncDir, "streams.json")
	writerPath := filepath.Join(syncDir, "writer.json")
	statePath := filepath.Join(syncDir, "state.json")

	// Write source config as config.json with proper permissions
	fmt.Printf("writing source config to %s\n", configPath)
	if err := writeFile(configPath, []byte(sourceConfig)); err != nil {
		return nil, fmt.Errorf("failed to write source config: %v", err)
	}

	// Write streams config as streams.json with proper permissions
	if err := writeFile(catalogPath, []byte(streamsConfig)); err != nil {
		return nil, fmt.Errorf("failed to write streams config: %v", err)
	}

	// Write destination config as writer.json with proper permissions
	if err := writeFile(writerPath, []byte(destConfig)); err != nil {
		return nil, fmt.Errorf("failed to write destination config: %v", err)
	}

	// Write state config as state.json with proper permissions
	if err := writeFile(statePath, []byte(stateConfig)); err != nil {
		return nil, fmt.Errorf("failed to write state config: %v", err)
	}

	// Execute sync command with additional arguments
	_, err = r.ExecuteDockerCommand("config", Sync, sourceType, version, configPath,
		"--catalog", "/mnt/config/streams.json",
		"--destination", "/mnt/config/writer.json",
		"--state", "/mnt/config/state.json")
	if err != nil {
		return nil, err
	}

	// List files in output directory for debugging
	r.ListOutputFiles(syncDir, "after sync")

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "completed",
			"message": "Sync completed successfully",
		}, nil
	}

	// Parse state file
	result, err := r.ParseJSONFile(statePath)
	if err != nil {
		return map[string]interface{}{
			"status": "completed",
			"error":  fmt.Sprintf("failed to parse state file: %v", err),
		}, nil
	}
	stateJSON, _ := json.Marshal(result)
	job.State = string(stateJSON)
	job.Active = true
	jobORM.Update(job)
	return result, nil
}
