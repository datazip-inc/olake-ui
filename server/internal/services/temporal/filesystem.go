package temporal

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"

	"github.com/datazip-inc/olake-ui/server/internal/storage"
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

// SetupConfigFiles writes the config files to the workflow directory before the worker runs, for
// direct-execution flows (discover, check, etc.).
func SetupConfigFiles(ctx context.Context, cmd Command, workflowID string, configs []JobConfig) error {
	files := make([]storage.JobConfig, 0, len(configs))
	for _, config := range configs {
		files = append(files, storage.JobConfig{RelativePath: config.Name, Data: config.Data})
	}
	if err := storage.WriteFiles(ctx, GetWorkflowDirectory(cmd, workflowID), files); err != nil {
		return fmt.Errorf("failed to write config files: %s", err)
	}
	return nil
}
