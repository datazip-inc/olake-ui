package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateDirectory creates a directory with the specified permissions if it doesn't exist
func CreateDirectory(dirPath string, perm os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %s", dirPath, err)
		}
	}
	return nil
}

// WriteFile writes data to a file, creating the directory if necessary
func WriteFile(filePath string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filePath)
	if err := CreateDirectory(dirPath, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, perm); err != nil {
		return fmt.Errorf("failed to write to file %s: %s", filePath, err)
	}
	return nil
}
