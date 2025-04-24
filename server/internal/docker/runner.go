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
	ConfigDir string
}

// NewRunner creates a new Docker runner
func NewRunner(configDir string) *Runner {
	return &Runner{
		ConfigDir: configDir,
	}
}

// GetDefaultConfigDir returns the default directory for storing config files
func GetDefaultConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/olake-config"
	}
	return filepath.Join(homeDir, ".olake", "config")
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
)

// WriteConfigToFile writes config JSON to a temporary file
func (r *Runner) WriteConfigToFile(config string, sourceID int) (string, error) {
	// Create directory if not exists
	if err := os.MkdirAll(r.ConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create a file path with sourceID to avoid conflicts
	configPath := filepath.Join(r.ConfigDir, fmt.Sprintf("config-%d.json", sourceID))

	// Write config to file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write config to file: %v", err)
	}

	return configPath, nil
}

// RunDockerCommand executes a Docker command for a source and returns the output
func (r *Runner) RunDockerCommand(sourceType, version, config string, sourceID int, cmd Command) (map[string]interface{}, error) {
	// Write config to file
	configPath, err := r.WriteConfigToFile(config, sourceID)
	if err != nil {
		return nil, err
	}

	// Create directory for output
	outputDir := filepath.Dir(configPath)

	// Define output file path that will be created by the command
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.json", "catalog"))

	// Construct Docker image name
	dockerImage := fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
	if version == "" {
		dockerImage = fmt.Sprintf("olakego/source-%s:latest", sourceType)
	}

	// Construct Docker command
	dockerCmd := exec.Command(
		"docker", "run", "--pull=always",
		"-v", fmt.Sprintf("%s:/mnt/config", outputDir),
		dockerImage,
		string(cmd),
		"--config", fmt.Sprintf("/mnt/config/%s", filepath.Base(configPath)),
	)

	// Execute Docker command
	output, err := dockerCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker command failed: %v, output: %s", err, string(output))
	}

	// Read the generated output file
	fileData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file %s: %v", outputPath, err)
	}

	// Parse JSON from file
	var result map[string]interface{}
	if err := json.Unmarshal(fileData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file: %v", err)
	}

	return result, nil
}

// GetCatalog runs the discover command and returns catalog data
func (r *Runner) GetCatalog(sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	return r.RunDockerCommand(sourceType, version, config, sourceID, Discover)
}

// GetSpec runs the spec command and returns connector specification
func (r *Runner) GetSpec(sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	return r.RunDockerCommand(sourceType, version, config, sourceID, Spec)
}

// TestConnection runs the check command and tests connection
func (r *Runner) TestConnection(sourceType, version, config string, sourceID int) (map[string]interface{}, error) {
	return r.RunDockerCommand(sourceType, version, config, sourceID, Check)
}
