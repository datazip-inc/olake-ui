package utils

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

type logChunkFile struct {
	index int
	name  string
}

// readLogsFromS3 reads the logs from S3 for a workflow.
func readLogsFromS3(ctx context.Context, workflowID string, cursor int64, _ int, direction string) (*dto.TaskLogsResponse, error) {
	workflowHash, err := GetAndValidateS3LogBaseDir(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	syncFolder, err := GetAndValidateS3SyncDir(ctx, workflowHash)
	if err != nil {
		return nil, err
	}

	syncLogDir := path.Join(workflowHash, "logs", syncFolder)
	chunks, err := listS3LogChunks(ctx, syncLogDir)
	if err != nil {
		return nil, err
	}

	return processS3Logs(ctx, syncLogDir, chunks, cursor, direction)
}

// GetAndValidateS3LogBaseDir validates the S3 log base directory.
func GetAndValidateS3LogBaseDir(ctx context.Context, workflowID string) (string, error) {
	workflowHash := WorkflowHash(workflowID)
	exists, err := storage.PrefixExists(ctx, workflowHash)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("logs directory not found: %s", workflowHash)
	}

	return workflowHash, nil
}

// GetAndValidateS3SyncDir gets the S3 sync folder name for a workflow.
func GetAndValidateS3SyncDir(ctx context.Context, workflowHash string) (string, error) {
	syncFolders, err := storage.ListFolderNamesWithPrefix(ctx, path.Join(workflowHash, "logs/sync_"))
	if err != nil {
		return "", err
	}
	if len(syncFolders) == 0 {
		return "", fmt.Errorf("no sync folder found in: %s/logs", workflowHash)
	}

	return syncFolders[0], nil
}

// listS3LogChunks lists the S3 log chunks for a workflow.
func listS3LogChunks(ctx context.Context, syncLogDir string) ([]logChunkFile, error) {
	names, err := storage.ListObjectNames(ctx, syncLogDir)
	if err != nil {
		return nil, err
	}

	chunks := make([]logChunkFile, 0, len(names))
	for _, name := range names {
		index, ok := parseNumberedLogChunkName(name, constants.ConnectorLogPrefix)
		if !ok {
			continue
		}
		chunks = append(chunks, logChunkFile{index: index, name: name})
	}

	slices.SortFunc(chunks, func(a, b logChunkFile) int {
		return a.index - b.index
	})
	return chunks, nil
}

// parseNumberedLogChunkName parses names like connector-000001-<timestamp>.log.
func parseNumberedLogChunkName(name, prefix string) (int, bool) {
	body := strings.TrimSuffix(name, ".log")
	remainder := body[len(prefix)+1:]
	dash := strings.Index(remainder, "-")
	if dash < 1 {
		return 0, false
	}

	index, err := strconv.Atoi(remainder[:dash])
	if err != nil {
		return 0, false
	}

	return index, true
}

// readS3ChunkLines reads the lines from a S3 log chunk.
func readS3ChunkLines(ctx context.Context, syncLogDir string, chunk logChunkFile) ([]string, error) {
	chunkPath := path.Join(syncLogDir, chunk.name)
	content, _, err := storage.ReadFileFromS3(ctx, "", chunkPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to read log chunk %s: %s", chunkPath, err)
	}

	lines := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if isValidLogLine(line) {
			lines = append(lines, line)
		}
	}

	return lines, nil
}

// processS3Logs processes the S3 logs for a workflow.
func processS3Logs(ctx context.Context, syncLogDir string, chunks []logChunkFile, cursor int64, direction string) (*dto.TaskLogsResponse, error) {
	dir := strings.ToLower(strings.TrimSpace(direction))
	if dir != "newer" {
		dir = constants.DefaultLogsDirection
	}

	var currentFileNumber int64
	switch {
	case cursor < 0:
		currentFileNumber = int64(chunks[len(chunks)-1].index)
	case dir == "older":
		// Older: cursor is the file number the UI already has.
		currentFileNumber = cursor - 1
	default:
		// Newer: cursor is the file number the UI wants next.
		currentFileNumber = cursor
	}

	firstFileNumber := int64(chunks[0].index)
	lastFileNumber := int64(chunks[len(chunks)-1].index)
	// Calculate the position of the chunk in the list.
	pos := int(currentFileNumber - firstFileNumber)
	if pos < 0 || pos >= len(chunks) {
		return &dto.TaskLogsResponse{Logs: []map[string]interface{}{}}, nil
	}

	// Read the lines from the chunk.
	lines, err := readS3ChunkLines(ctx, syncLogDir, chunks[pos])
	if err != nil {
		return nil, err
	}

	return &dto.TaskLogsResponse{
		Logs:         parseLines(lines),
		OlderCursor:  currentFileNumber,
		NewerCursor:  currentFileNumber + 1,
		HasMoreOlder: currentFileNumber > firstFileNumber,
		HasMoreNewer: currentFileNumber < lastFileNumber,
	}, nil
}

func addS3FilesToArchive(ctx context.Context, workflowHash string, tarWriter *tar.Writer) error {
	stateFile := path.Join(workflowHash, "state.json")
	body, modTime, err := storage.ReadFileFromS3(ctx, "", stateFile, false)
	if err != nil {
		logger.Warnf("failed to add state.json to archive: %s", err)
	} else if err := addBytesToArchive(tarWriter, "state.json", []byte(body), modTime); err != nil {
		return err
	}

	objectPaths, err := storage.ListAllObjectRelPaths(ctx, path.Join(workflowHash, "logs"))
	if err != nil {
		return err
	}

	for _, objectPath := range objectPaths {
		archiveName, ok := archiveNameUnderWorkflow(workflowHash, objectPath)
		if !ok {
			continue
		}

		body, modTime, err := storage.ReadFileFromS3(ctx, "", objectPath, false)
		if err != nil {
			logger.Warnf("failed to add %s to archive: %s", objectPath, err)
			continue
		}

		if err := addBytesToArchive(tarWriter, archiveName, []byte(body), modTime); err != nil {
			return err
		}
	}

	return nil
}

func archiveNameUnderWorkflow(workflowHash, objectPath string) (string, bool) {
	prefix := strings.TrimSuffix(workflowHash, "/") + "/"
	name := strings.TrimPrefix(objectPath, prefix)
	if strings.HasPrefix(name, "logs/") {
		return path.Join("logs", path.Base(name)), true
	}

	return "", false
}

func addBytesToArchive(tarWriter *tar.Writer, nameInArchive string, data []byte, modTime time.Time) error {
	header := &tar.Header{
		Name:    nameInArchive,
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: modTime,
	}

	_, err := writeToArchive(tarWriter, header, bytes.NewReader(data))
	return err
}
