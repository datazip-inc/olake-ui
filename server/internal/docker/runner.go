package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner is responsible for executing Docker commands
type Runner struct {
	WorkingDir string
}

// NewRunner creates a new Docker runner
func NewRunner(workingDir string) *Runner {
	return &Runner{
		WorkingDir: workingDir,
	}
}

// GetDefaultConfigDir returns the default directory for storing config files
func GetDefaultConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/olake-config"
	}
	return filepath.Join(homeDir, ".olake")
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

// WriteToFile writes data to a file with a specific name based on ID
func (r *Runner) WriteToFile(fileData string, ID any) (string, error) {
	// Create directory if not exists
	if err := os.MkdirAll(r.WorkingDir, 0755); err != nil {
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

	// Write config to file
	if err := os.WriteFile(configPath, []byte(fileData), 0644); err != nil {
		return "", fmt.Errorf("failed to write config to file: %v", err)
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
func (r *Runner) ExecuteDockerCommand(command Command, sourceType, version, configPath string, additionalArgs ...string) ([]byte, error) {
	outputDir := filepath.Dir(configPath)

	// Construct Docker command arguments
	dockerArgs := []string{
		"run", "--pull=always",
		"-v", fmt.Sprintf("%s:/mnt/config", outputDir),
		r.GetDockerImageName(sourceType, version),
		string(command),
		"--config", fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)),
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
				err := os.WriteFile(targetPath, tryData, 0644)
				if err == nil {
					return tryData, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no suitable JSON file found")
}

// GetCatalog runs the discover command and returns catalog data
func (r *Runner) GetCatalog(sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	// Write config to file
	configPath, err := r.WriteToFile(config, sourceID)
	if err != nil {
		return nil, err
	}

	// Define catalog output path
	outputDir := filepath.Dir(configPath)
	catalogPath := filepath.Join(outputDir, "streams.json")

	// Execute discover command
	_, err = r.ExecuteDockerCommand(Discover, sourceType, version, configPath)
	if err != nil {
		return nil, err
	}

	// List files in output directory for debugging
	r.ListOutputFiles(outputDir, "after discover")

	// Check if catalog file exists
	_, err = os.ReadFile(catalogPath)
	if err != nil {
		// Try to find any JSON file that might be the catalog
		_, err = r.FindAlternativeJSONFile(outputDir, catalogPath, configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read catalog file: %v", err)
		}
	}

	// Parse JSON from file
	return r.ParseJSONFile(catalogPath)
}

// RunSync runs the sync command to transfer data from source to destination
func (r *Runner) RunSync(sourceType, version, sourceConfig, destConfig, streamsConfig string, JobId, projectID, sourceID, destID int, workflowID string) (map[string]interface{}, string, error) {
	// Create sync folder
	syncFolderName := workflowID
	// Create directory for output
	syncDir := filepath.Join(r.WorkingDir, syncFolderName)
	fmt.Printf("working directory path %s\n", syncDir)

	// Ensure the directory exists
	if err := os.MkdirAll(syncDir, 0755); err != nil {
		return nil, syncDir, fmt.Errorf("failed to create sync directory: %v", err)
	}

	// Define paths for required files
	configPath := filepath.Join(syncDir, "config.json")
	catalogPath := filepath.Join(syncDir, "streams.json")
	writerPath := filepath.Join(syncDir, "writer.json")
	statePath := filepath.Join(syncDir, "state.json")

	// Set permissions for the output directory
	cmd := exec.Command("sudo", "chmod", "-R", "777", syncDir)
	_ = cmd.Run() // Ignore error; permission setting is not critical
	// write source config as config.json
	fmt.Printf("writing source config to %s\n", configPath)
	err := os.WriteFile(configPath, []byte(sourceConfig), 0755)
	if err != nil {
		return nil, syncDir, fmt.Errorf("failed to write source config: %v", err)
	}
	// Write streams config as streams.json
	err = os.WriteFile(catalogPath, []byte(streamsConfig), 0755)
	if err != nil {
		return nil, syncDir, fmt.Errorf("failed to write streams config: %v", err)
	}
	// Write destination config as writer.json
	err = os.WriteFile(writerPath, []byte(destConfig), 0755)
	if err != nil {
		return nil, syncDir, fmt.Errorf("failed to write destination config: %v", err)
	}

	// Execute sync command with additional arguments
	_, err = r.ExecuteDockerCommand(Sync, sourceType, version, configPath,
		"--catalog", "/mnt/config/streams.json",
		"--destination", "/mnt/config/writer.json")
	if err != nil {
		return nil, syncDir, err
	}

	// List files in output directory for debugging
	r.ListOutputFiles(syncDir, "after sync")

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "completed",
			"message": "Sync completed successfully",
		}, syncDir, nil
	}

	// Parse state file
	result, err := r.ParseJSONFile(statePath)
	if err != nil {
		return map[string]interface{}{
			"status": "completed",
			"error":  fmt.Sprintf("failed to parse state file: %v", err),
		}, syncDir, nil
	}

	return result, syncDir, nil
}
