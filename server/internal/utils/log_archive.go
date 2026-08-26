package utils

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// AddBytesToArchive writes a byte slice as a file entry in a tar archive.
func AddBytesToArchive(tarWriter *tar.Writer, nameInArchive string, data []byte) error {
	header := &tar.Header{
		Name:    nameInArchive,
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %s", nameInArchive, err)
	}

	if _, err := tarWriter.Write(data); err != nil {
		return fmt.Errorf("failed to write tar content for %s: %s", nameInArchive, err)
	}

	return nil
}

// StreamTaskLogArchive streams state.json and the full logs/ tree (connector + worker) as tar.gz.
func StreamTaskLogArchive(ctx context.Context, workflowID string, writer io.Writer) error {
	mode := strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode))
	if mode == constants.StorageModeS3 {
		return streamTaskLogArchiveS3(ctx, workflowID, writer)
	}

	baseDir, err := GetAndValidateLogBaseDir(workflowID)
	if err != nil {
		return err
	}

	return streamTaskLogArchiveNFS(baseDir, writer)
}

func streamTaskLogArchiveNFS(baseDir string, writer io.Writer) error {
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	stateFile := filepath.Join(baseDir, "state.json")
	if err := AddFileToArchive(tarWriter, stateFile, "state.json"); err != nil {
		logger.Warnf("failed to add state.json to archive: %s", err)
	}

	logsRoot := filepath.Join(baseDir, "logs")
	if _, err := os.Stat(logsRoot); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(logsRoot, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(logsRoot, filePath)
		if err != nil {
			return err
		}

		archivePath := filepath.Join("logs", rel)
		return AddFileToArchive(tarWriter, filePath, archivePath)
	})
}

func streamTaskLogArchiveS3(ctx context.Context, workflowID string, writer io.Writer) error {
	workflowHash, err := validateS3LogBase(ctx, workflowID)
	if err != nil {
		return err
	}

	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	stateRel := path.Join(workflowHash, "state.json")
	if body, err := storage.ReadFileFromS3(ctx, "", stateRel, false); err == nil {
		if err := AddBytesToArchive(tarWriter, "state.json", []byte(body)); err != nil {
			return err
		}
	} else {
		logger.Warnf("failed to add state.json to archive: %s", err)
	}

	objectPaths, err := storage.ListAllObjectRelPaths(ctx, path.Join(workflowHash, "logs"))
	if err != nil {
		return err
	}

	for _, relPath := range objectPaths {
		archiveName, ok := archiveNameUnderWorkflow(workflowHash, relPath)
		if !ok {
			continue
		}

		body, err := storage.ReadFileFromS3(ctx, "", relPath, false)
		if err != nil {
			logger.Warnf("failed to add %s to archive: %s", relPath, err)
			continue
		}

		if err := AddBytesToArchive(tarWriter, archiveName, []byte(body)); err != nil {
			return err
		}
	}

	return nil
}

func archiveNameUnderWorkflow(workflowHash, relPath string) (string, bool) {
	prefix := strings.TrimSuffix(workflowHash, "/") + "/"
	if !strings.HasPrefix(relPath, prefix) {
		return "", false
	}

	name := strings.TrimPrefix(relPath, prefix)
	if name == "state.json" {
		return "state.json", true
	}
	if strings.HasPrefix(name, "logs/") {
		return name, true
	}

	return "", false
}
