package temporal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
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
		if err := utils.WriteFile(filePath, []byte(config.Data), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %s", config.Name, err)
		}
	}
	return nil
}

func readJSONFileFromNFS(filePath string) (map[string]interface{}, error) {
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

// SetupConfigFiles creates the work directory and writes the config files to it.
// For direct-execution flows (discover, check, etc.) the BFF writes configs before
// the worker runs. NFS writes to the shared mount; S3 uploads to the same object
// keys used in worker command paths.
func SetupConfigFiles(ctx context.Context, cmd Command, workflowID string, configs []JobConfig) error {
	switch strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode)) {
	case constants.StorageModeS3:
		return writeConfigFilesToS3(ctx, cmd, workflowID, configs)
	case constants.StorageModeNFS:
		subDir := GetWorkflowDirectory(cmd, workflowID)
		workDir := filepath.Join(constants.DefaultConfigDir, subDir)

		if err := createDirectory(workDir, 0755); err != nil {
			return fmt.Errorf("failed to create work directory %s: %s", workDir, err)
		}

		if err := writeConfigFiles(workDir, configs); err != nil {
			return fmt.Errorf("failed to write config files: %s", err)
		}

		return nil
	}

	return fmt.Errorf("unsupported OLAKE_STORAGE_MODE: %q", appconfig.Load().OlakeStorageMode)
}
