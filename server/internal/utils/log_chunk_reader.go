package utils

import (
	"context"
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

type logChunkFile struct {
	index int
	name  string
}

// NormalizeLogType returns connector (default) or worker.
func NormalizeLogType(logType string) string {
	switch strings.ToLower(strings.TrimSpace(logType)) {
	case constants.LogTypeWorker:
		return constants.LogTypeWorker
	default:
		return constants.LogTypeConnector
	}
}

func logPrefixForType(logType string) string {
	if logType == constants.LogTypeWorker {
		return constants.WorkerLogPrefix
	}
	return constants.ConnectorLogPrefix
}

// parseNumberedLogChunkName parses names like connector-000001-<timestamp>.log or worker-000002-<timestamp>.log.
func parseNumberedLogChunkName(name, prefix string) (int, bool) {
	if !strings.HasPrefix(name, prefix+"-") || !strings.HasSuffix(name, ".log") {
		return 0, false
	}

	body := strings.TrimSuffix(name, ".log")
	remainder := body[len(prefix)+1:]
	dash := strings.Index(remainder, "-")
	if dash < 1 {
		return 0, false
	}

	seq := remainder[:dash]
	for _, c := range seq {
		if c < '0' || c > '9' {
			return 0, false
		}
	}

	index, err := strconv.Atoi(seq)
	if err != nil {
		return 0, false
	}

	return index, true
}

func sortLogChunks(chunks []logChunkFile) []logChunkFile {
	out := make([]logChunkFile, len(chunks))
	copy(out, chunks)
	slices.SortFunc(out, func(a, b logChunkFile) int {
		return a.index - b.index
	})
	return out
}

func filterAndSortLogChunkNames(names []string, prefix string) ([]logChunkFile, error) {
	chunks := make([]logChunkFile, 0, len(names))
	for _, name := range names {
		index, ok := parseNumberedLogChunkName(name, prefix)
		if !ok {
			continue
		}
		chunks = append(chunks, logChunkFile{index: index, name: name})
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no %s log chunks found", prefix)
	}

	return sortLogChunks(chunks), nil
}

func listS3LogChunks(ctx context.Context, logDirRel, prefix string) ([]logChunkFile, error) {
	names, err := storage.ListObjectNames(ctx, logDirRel)
	if err != nil {
		return nil, err
	}
	return filterAndSortLogChunkNames(names, prefix)
}

func readS3ChunkLines(ctx context.Context, logDirRel string, chunk logChunkFile) ([]string, error) {
	relPath := path.Join(logDirRel, chunk.name)
	content, err := storage.ReadObjectBytes(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log chunk %s: %s", relPath, err)
	}
	return parseValidLogLinesFromBytes(content), nil
}

// paginateLogChunks applies client-driven pagination. Cursors are chunk indices managed by the UI.
func paginateLogChunks(
	chunks []logChunkFile,
	cursor int64,
	direction string,
	readChunk func(chunk logChunkFile) ([]string, error),
) (*dto.TaskLogsResponse, error) {
	numChunks := int64(len(chunks))
	chunkLimit := int64(constants.S3LogChunksPerRequest)

	dir := strings.ToLower(strings.TrimSpace(direction))
	if dir != "newer" {
		dir = constants.DefaultLogsDirection
	}

	isTail := cursor < 0
	response := &dto.TaskLogsResponse{}

	if isTail || dir == "older" {
		newerBound := numChunks
		endExclusive := numChunks
		if !isTail {
			if cursor > numChunks {
				cursor = numChunks
			}
			newerBound = cursor
			endExclusive = cursor
		}

		if endExclusive <= 0 {
			response.Logs = []map[string]interface{}{}
			response.OlderCursor = 0
			response.NewerCursor = 0
			response.HasMoreOlder = false
			response.HasMoreNewer = false
			return response, nil
		}

		start := max(endExclusive-chunkLimit, 0)
		lines, err := readChunkRange(chunks, start, endExclusive, readChunk)
		if err != nil {
			return nil, err
		}

		response.Logs = parseLogLines(lines)
		response.OlderCursor = start
		response.NewerCursor = newerBound
		response.HasMoreOlder = start > 0
		response.HasMoreNewer = newerBound < numChunks
		return response, nil
	}

	startInclusive := cursor
	if startInclusive < 0 {
		startInclusive = 0
	}
	if startInclusive > numChunks {
		startInclusive = numChunks
	}

	if startInclusive >= numChunks {
		response.Logs = []map[string]interface{}{}
		response.OlderCursor = numChunks
		response.NewerCursor = numChunks
		response.HasMoreOlder = numChunks > 0
		response.HasMoreNewer = false
		return response, nil
	}

	endExclusive := min(startInclusive+chunkLimit, numChunks)
	lines, err := readChunkRange(chunks, startInclusive, endExclusive, readChunk)
	if err != nil {
		return nil, err
	}

	response.Logs = parseLogLines(lines)
	response.OlderCursor = startInclusive
	response.NewerCursor = endExclusive
	response.HasMoreNewer = endExclusive < numChunks
	response.HasMoreOlder = startInclusive > 0

	return response, nil
}

func readChunkRange(
	chunks []logChunkFile,
	startInclusive, endExclusive int64,
	readChunk func(chunk logChunkFile) ([]string, error),
) ([]string, error) {
	lines := make([]string, 0)
	for i := startInclusive; i < endExclusive; i++ {
		chunkLines, err := readChunk(chunks[i])
		if err != nil {
			return nil, err
		}
		lines = append(lines, chunkLines...)
	}
	return lines, nil
}
