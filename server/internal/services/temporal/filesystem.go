package temporal

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

var (
	AsyncCommands = []Command{Sync, ClearDestination}
)

// getWorkflowDirectory determines the directory name based on operation and workflow ID
func getWorkflowDirectory(operation Command, originalWorkflowID string) string {
	if slices.Contains(AsyncCommands, operation) {
		return fmt.Sprintf("%x", sha256.Sum256([]byte(originalWorkflowID)))
	}
	return originalWorkflowID
}

// createDirectory creates a directory with the specified permissions if it doesn't exist
func createDirectory(dirPath string, perm os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %s", dirPath, err)
		}
	}
	return nil
}

// writeConfigFiles writes the config files to the work directory
func writeConfigFiles(workDir string, configs []JobConfig) error {
	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := os.WriteFile(filePath, []byte(config.Data), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %s", config.Name, err)
		}
	}
	return nil
}

func ReadJSONFile(filePath string) (map[string]interface{}, error) {
	fileOutput, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(fileOutput, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %v", err)
	}

	return result, nil
}

// SetupConfigFiles creates the work directory and writes the config files to it
// It writes to the mounted path and can be accessed by the worker.
func SetupConfigFiles(cmd Command, workflowID string, configs []JobConfig) error {
	subDir := getWorkflowDirectory(cmd, workflowID)
	workDir := filepath.Join(constants.DefaultConfigDir, subDir)

	if err := createDirectory(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory %s: %s", workDir, err)
	}

	if err := writeConfigFiles(workDir, configs); err != nil {
		return fmt.Errorf("failed to write config files: %s", err)
	}

	return nil
}
