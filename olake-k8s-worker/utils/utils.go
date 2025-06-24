package utils

import (
	"os"
)

// CreateDirectory creates a directory with the specified permissions if it doesn't exist
func CreateDirectory(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// WriteFile writes data to a file with the specified permissions
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}
