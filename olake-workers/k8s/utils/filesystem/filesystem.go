package filesystem

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"olake-ui/olake-workers/k8s/shared"
)

const minStateFileSize = 10

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

// ReadAndValidateStateFile reads and validates the state.json file for the given workflow.
// Returns the raw file contents as []byte if the file exists and is valid JSON.
// Returns os.ErrNotExist if the file does not exist.
func (fs *Helper) ReadAndValidateStateFile(workflowID string) ([]byte, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflowID cannot be empty")
	}

	workflowDir := fs.GetWorkflowDirectory(shared.Sync, workflowID)
	statePath := fs.GetFilePath(workflowDir, "state.json")

	stateData, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", os.ErrNotExist, statePath)
		}
		return nil, fmt.Errorf("read failed: %w", err)
	}

	// Validate file size first (cheaper than JSON parsing)
	if len(stateData) < minStateFileSize {
		return nil, fmt.Errorf("state file too small (%d bytes)", len(stateData))
	}

	// Validate JSON structure
	var js json.RawMessage
	if err := json.Unmarshal(stateData, &js); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return stateData, nil
}
