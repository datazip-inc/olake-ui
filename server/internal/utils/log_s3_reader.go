package utils

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

func validateS3LogBase(ctx context.Context, workflowID string) (string, error) {
	if workflowID == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	baseRel := GetLogBaseRelPath(workflowID)
	exists, err := storage.PrefixExists(ctx, baseRel)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("logs directory not found: %s", baseRel)
	}

	return baseRel, nil
}

func getS3SyncFolder(ctx context.Context, baseRel string) (string, error) {
	children, err := storage.ListCommonPrefixes(ctx, path.Join(baseRel, "logs"))
	if err != nil {
		return "", err
	}

	var syncFolders []string
	for _, name := range children {
		if strings.HasPrefix(name, "sync_") {
			syncFolders = append(syncFolders, name)
		}
	}
	if len(syncFolders) == 0 {
		// Worker logs may exist before the first connector chunk is uploaded.
		return "", fmt.Errorf("no sync folder found in: %s/logs", baseRel)
	}

	// Prefer the latest session when multiple sync_* folders exist.
	sort.Strings(syncFolders)
	return syncFolders[len(syncFolders)-1], nil
}

func emptyTaskLogsResponse() *dto.TaskLogsResponse {
	return &dto.TaskLogsResponse{
		Logs:         []map[string]interface{}{},
		OlderCursor:  0,
		NewerCursor:  0,
		HasMoreOlder: false,
		HasMoreNewer: false,
	}
}

func resolveS3LogDirRel(baseRel, syncFolder, logType string) (string, error) {
	switch logType {
	case constants.LogTypeWorker:
		return path.Join(baseRel, "logs", constants.WorkerLogsDir), nil
	case constants.LogTypeConnector:
		if syncFolder == "" {
			return "", fmt.Errorf("sync folder is required for connector logs")
		}
		return path.Join(baseRel, "logs", syncFolder), nil
	default:
		return "", fmt.Errorf("unsupported log type: %s", logType)
	}
}

func readLogsFromS3(ctx context.Context, workflowID, logType string, cursor int64, _ int, direction string) (*dto.TaskLogsResponse, error) {
	logType = NormalizeLogType(logType)
	prefix := logPrefixForType(logType)

	baseRel, err := validateS3LogBase(ctx, workflowID)
	if err != nil {
		if strings.Contains(err.Error(), "logs directory not found") {
			return emptyTaskLogsResponse(), nil
		}
		return nil, err
	}

	var syncFolder string
	if logType == constants.LogTypeConnector {
		syncFolder, err = getS3SyncFolder(ctx, baseRel)
		if err != nil {
			if strings.Contains(err.Error(), "no sync folder found") {
				// Connector chunks may not have been uploaded yet (only worker/ exists).
				return emptyTaskLogsResponse(), nil
			}
			return nil, err
		}
	}

	logDirRel, err := resolveS3LogDirRel(baseRel, syncFolder, logType)
	if err != nil {
		return nil, err
	}

	chunks, err := listS3LogChunks(ctx, logDirRel, prefix)
	if err != nil {
		if strings.Contains(err.Error(), "no "+prefix+" log chunks found") {
			return emptyTaskLogsResponse(), nil
		}
		return nil, err
	}

	return paginateLogChunks(chunks, cursor, direction, func(chunk logChunkFile) ([]string, error) {
		return readS3ChunkLines(ctx, logDirRel, chunk)
	})
}

func ReadTaskLogs(ctx context.Context, workflowID, logType string, cursor int64, limit int, direction string) (*dto.TaskLogsResponse, error) {
	logType = NormalizeLogType(logType)

	mode := strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode))
	switch mode {
	case constants.StorageModeS3:
		return readLogsFromS3(ctx, workflowID, logType, cursor, limit, direction)
	default:
		mainSyncDir, err := GetAndValidateLogBaseDir(workflowID)
		if err != nil {
			return nil, err
		}
		return readLogsFromNFS(mainSyncDir, logType, cursor, limit, direction)
	}
}

func readLogsFromNFS(mainSyncDir, logType string, cursor int64, limit int, direction string) (*dto.TaskLogsResponse, error) {
	switch logType {
	case constants.LogTypeWorker:
		workerLogPath := filepath.Join(mainSyncDir, "logs", constants.WorkerLogFile)
		return ReadLogsFromFile(workerLogPath, cursor, limit, direction)
	case constants.LogTypeConnector:
		return ReadLogsFromDir(mainSyncDir, cursor, limit, direction)
	default:
		return nil, fmt.Errorf("unsupported log type: %s", logType)
	}
}
