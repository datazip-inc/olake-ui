package utils

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

// LogEntry represents a log entry
type LogEntry struct {
	Level   string          `json:"level"`
	Time    time.Time       `json:"time"`
	Message json.RawMessage `json:"message"` // store raw JSON
}

func ReadLogs(ctx context.Context, logDir string, cursor int64, limit int, direction string) (*dto.TaskLogsResponse, error) {
	switch storage.Mode() {
	case constants.StorageModeS3:
		if workRel, err := filepath.Rel(constants.DefaultConfigDir, logDir); err == nil {
			logDir = workRel
		}
		return readLogsFromS3(ctx, logDir, cursor, direction)
	default:
		return readLogsFromFilesystem(logDir, cursor, limit, direction)
	}
}

// GetLogArchiveFilename generates the filename for the log archive download.
func GetLogArchiveFilename(ctx context.Context, jobID int, filePath string) (string, error) {
	baseDir, err := GetAndValidateLogBaseDir(ctx, filePath)
	if err != nil {
		return "", err
	}

	syncFolder, err := GetAndValidateSyncFolder(ctx, baseDir)
	if err != nil {
		return "", err
	}

	syncTimestamp := strings.ReplaceAll(strings.TrimPrefix(syncFolder, "sync_"), "_", "-")
	return fmt.Sprintf("job-%d-logs-%s.tar.gz", jobID, syncTimestamp), nil
}

// GetAndValidateLogBaseDir returns the log base path for the active storage mode.
func GetAndValidateLogBaseDir(ctx context.Context, workflowID string) (string, error) {
	switch storage.Mode() {
	case constants.StorageModeS3:
		return GetAndValidateS3LogBaseDir(ctx, workflowID)
	default:
		return GetAndValidateFilesystemLogBaseDir(workflowID)
	}
}

// GetAndValidateSyncFolder returns the sync_* folder name for the active storage mode.
func GetAndValidateSyncFolder(ctx context.Context, baseDir string) (string, error) {
	switch storage.Mode() {
	case constants.StorageModeS3:
		return GetAndValidateS3SyncDir(ctx, baseDir)
	default:
		_, syncFolderName, err := GetAndValidateFilesystemSyncDir(baseDir)
		return syncFolderName, err
	}
}

// parseValidLogLine parses a line and reports whether it is a valid, non-debug log entry.
// Returns the parsed entry so callers can reuse it instead of unmarshalling the line again.
func parseValidLogLine(line string) (LogEntry, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return LogEntry{}, false
	}

	var logEntry LogEntry
	if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
		return LogEntry{}, false
	}

	if logEntry.Level == "debug" {
		return LogEntry{}, false
	}

	return logEntry, true
}

func parseLines(entries []LogEntry) []map[string]interface{} {
	batch := make([]map[string]interface{}, 0, len(entries))
	for _, logEntry := range entries {
		var messageStr string
		var tmp interface{}
		if err := json.Unmarshal(logEntry.Message, &tmp); err == nil {
			switch v := tmp.(type) {
			case string:
				messageStr = v
			default:
				msgBytes, err := json.Marshal(v)
				if err != nil {
					messageStr = string(logEntry.Message)
				} else {
					messageStr = string(msgBytes)
				}
			}
		} else {
			messageStr = string(logEntry.Message)
		}

		batch = append(batch, map[string]interface{}{
			"level":   logEntry.Level,
			"time":    logEntry.Time.UTC().Format(time.RFC3339),
			"message": messageStr,
		})
	}

	return batch
}

// AddFilesToArchive adds state.json and the logs/ tree to the tar archive for the active storage mode.
func AddFilesToArchive(ctx context.Context, baseDir string, tarWriter *tar.Writer) error {
	switch storage.Mode() {
	case constants.StorageModeS3:
		return addS3FilesToArchive(ctx, baseDir, tarWriter)
	default:
		return addFilesystemFilesToArchive(baseDir, tarWriter)
	}
}

func writeToArchive(tarWriter *tar.Writer, header *tar.Header, reader io.Reader) (int64, error) {
	if err := tarWriter.WriteHeader(header); err != nil {
		return 0, fmt.Errorf("failed to write tar header for %s: %s", header.Name, err)
	}

	bytesWritten, err := io.Copy(tarWriter, reader)
	if err != nil {
		return 0, fmt.Errorf("failed to write file content for %s: %s", header.Name, err)
	}

	return bytesWritten, nil
}
