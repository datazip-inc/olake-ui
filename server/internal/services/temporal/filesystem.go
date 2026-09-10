package temporal

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

var (
	AsyncCommands = []Command{Sync, ClearDestination}
)

// GetWorkflowDirectory determines the directory name based on operation and workflow ID
func GetWorkflowDirectory(operation Command, originalWorkflowID string) string {
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
		if err := utils.WriteFile(filePath, []byte(config.Data), constants.DefaultFileMode); err != nil {
			return fmt.Errorf("failed to write %s: %s", config.Name, err)
		}
	}
	return nil
}

// SetupConfigFiles creates the work directory and writes the config files to it
// It writes to the mounted path and can be accessed by the worker.
func SetupConfigFiles(cmd Command, workflowID string, configs []JobConfig) error {
	subDir := GetWorkflowDirectory(cmd, workflowID)
	workDir := filepath.Join(constants.DefaultConfigDir, subDir)

	if err := createDirectory(workDir, constants.DefaultDirMode); err != nil {
		return fmt.Errorf("failed to create work directory %s: %s", workDir, err)
	}

	if err := writeConfigFiles(workDir, configs); err != nil {
		return fmt.Errorf("failed to write config files: %s", err)
	}

	return nil
}
