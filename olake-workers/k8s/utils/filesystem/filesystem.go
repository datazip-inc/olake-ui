package filesystem

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"olake-ui/olake-workers/k8s/shared"
)

// Helper handles file system operations for job configuration
type Helper struct {
	basePath string
}

// NewHelper creates a new filesystem helper
func NewHelper() *Helper {
	return &Helper{
		basePath: "/data/olake-jobs", // PV mount point on worker pod
	}
}

// GetWorkflowDirectory determines the directory name based on operation and workflow ID
func (fs *Helper) GetWorkflowDirectory(operation shared.Command, originalWorkflowID string) string {
	if operation == shared.Sync {
		// Sync: use SHA256 hash (like Docker does)
		return fmt.Sprintf("%x", sha256.Sum256([]byte(originalWorkflowID)))
	} else {
		// Test/Discover: use WorkflowID directly (like Docker does)
		return originalWorkflowID
	}
}

// SetupWorkDirectory creates the work directory for a workflow
func (fs *Helper) SetupWorkDirectory(workflowDir string) error {
	workDir := filepath.Join(fs.basePath, workflowDir)
	return createDirectory(workDir, 0755)
}

// WriteConfigFiles writes configuration files to the workflow directory
func (fs *Helper) WriteConfigFiles(workflowDir string, configs []shared.JobConfig) error {
	workDir := filepath.Join(fs.basePath, workflowDir)

	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := writeFile(filePath, []byte(config.Data), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", config.Name, err)
		}
	}
	return nil
}

// GetFilePath returns the full path to a file in the workflow directory
func (fs *Helper) GetFilePath(workflowDir, fileName string) string {
	return filepath.Join(fs.basePath, workflowDir, fileName)
}

// Private helper functions for filesystem operations
func createDirectory(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}
