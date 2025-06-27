package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"olake-k8s-worker/shared"
)

// FilesystemHelper handles file system operations for job configuration
type FilesystemHelper struct {
	basePath string
}

// NewFilesystemHelper creates a new filesystem helper
func NewFilesystemHelper() *FilesystemHelper {
	return &FilesystemHelper{
		basePath: "/data/olake-jobs", // PV mount point on worker pod
	}
}

// NewFilesystemHelperWithPath creates a new filesystem helper with custom base path
func NewFilesystemHelperWithPath(basePath string) *FilesystemHelper {
	return &FilesystemHelper{
		basePath: basePath,
	}
}

// GetWorkflowDirectory determines the directory name based on operation and workflow ID
func (fs *FilesystemHelper) GetWorkflowDirectory(operation shared.Command, originalWorkflowID string) string {
	if operation == shared.Sync {
		// Sync: use SHA256 hash (like Docker does)
		return fmt.Sprintf("%x", sha256.Sum256([]byte(originalWorkflowID)))
	} else {
		// Test/Discover: use WorkflowID directly (like Docker does)
		return originalWorkflowID
	}
}

// SetupWorkDirectory creates the work directory for a workflow
func (fs *FilesystemHelper) SetupWorkDirectory(workflowDir string) error {
	workDir := filepath.Join(fs.basePath, workflowDir)
	return createDirectory(workDir, 0755)
}

// WriteConfigFiles writes configuration files to the workflow directory
func (fs *FilesystemHelper) WriteConfigFiles(workflowDir string, configs []shared.JobConfig) error {
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
func (fs *FilesystemHelper) GetFilePath(workflowDir, fileName string) string {
	return filepath.Join(fs.basePath, workflowDir, fileName)
}

// Private helper functions for filesystem operations
func createDirectory(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}
