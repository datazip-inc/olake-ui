package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// WriteFiles writes files under relDir, a path relative to the config dir shared with the
// worker: the NFS mount, or the matching S3 prefix.
func WriteFiles(ctx context.Context, relDir string, files []JobConfig) error {
	workDir := filepath.Join(constants.DefaultConfigDir, relDir)
	if Mode() == constants.StorageModeS3 {
		return WriteFilesToS3(ctx, workDir, files)
	}

	for _, file := range files {
		filePath := filepath.Join(workDir, file.RelativePath)
		if err := os.MkdirAll(filepath.Dir(filePath), constants.DefaultDirMode); err != nil {
			return fmt.Errorf("failed to create directory for %s: %s", filePath, err)
		}
		if err := os.WriteFile(filePath, []byte(file.Data), constants.DefaultFileMode); err != nil {
			return fmt.Errorf("failed to write %s: %s", filePath, err)
		}
	}
	return nil
}

// ReadFile reads relPath, a path relative to the config dir shared with the worker. A missing
// file returns an error that matches fs.ErrNotExist in both storage modes.
func ReadFile(ctx context.Context, relPath string) ([]byte, error) {
	if Mode() == constants.StorageModeS3 {
		body, _, err := ReadFileFromS3(ctx, "", relPath, false)
		if err != nil {
			return nil, err
		}
		return []byte(body), nil
	}

	data, err := os.ReadFile(filepath.Join(constants.DefaultConfigDir, relPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", relPath, err)
	}
	return data, nil
}
