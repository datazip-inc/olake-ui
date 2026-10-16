package utils

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

// openS3TaskLogs opens the chunk series for a source, in merge order (sync
// first): the connector's chunks in logs/sync_*/ and the worker's in logs/worker/.
func openS3TaskLogs(ctx context.Context, workflowDir, source string) ([]logFile, error) {
	var files []logFile
	if source == LogSourceAll || source == LogSourceSync {
		syncFolder, err := GetAndValidateS3SyncDir(ctx, workflowDir)
		switch {
		case err == nil:
			f, err := openS3LogSeries(ctx, LogSourceSync, path.Join(workflowDir, "logs", syncFolder), constants.ConnectorLogPrefix)
			if err != nil {
				return nil, err
			}
			files = append(files, f)
		// connector chunks are not uploaded until the first log flush
		case !strings.Contains(err.Error(), "no sync folder found"):
			return nil, err
		}
	}
	if source == LogSourceAll || source == LogSourceWorker {
		f, err := openS3LogSeries(ctx, LogSourceWorker, path.Join(workflowDir, constants.WorkerLogRelDir), constants.WorkerLogPrefix)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// s3LogSeries reads one source's numbered log chunks as a single log.
// A line's position is its chunk number in the high 32 bits and its line
// number within the chunk in the low 32 bits. Chunks are written once, so a
// position stays valid while new chunks are added.
type s3LogSeries struct {
	ctx    context.Context
	dir    string
	chunks []logChunkFile // ascending by chunk number
}

func s3LinePos(chunk, line int) int64 {
	return int64(chunk)<<32 | int64(line)
}

func openS3LogSeries(ctx context.Context, source, dir, prefix string) (logFile, error) {
	chunks, err := listS3LogChunks(ctx, dir, prefix)
	if err != nil {
		return logFile{}, err
	}
	s := &s3LogSeries{ctx: ctx, dir: dir, chunks: chunks}

	// the end is the start of the next chunk, so no chunk is downloaded here
	var size int64
	if len(chunks) > 0 {
		size = s3LinePos(chunks[len(chunks)-1].index+1, 0)
	}
	return logFile{source: source, size: size, readOlder: s.readOlder, readNewer: s.readNewer}, nil
}

func (s *s3LogSeries) chunkLines(chunk logChunkFile) ([]string, error) {
	chunkPath := path.Join(s.dir, chunk.name)
	content, _, err := storage.ReadFileFromS3(s.ctx, "", chunkPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to read log chunk %s: %s", chunkPath, err)
	}
	return strings.Split(strings.TrimRight(content, "\n"), "\n"), nil
}

func (s *s3LogSeries) readOlder(start int64, limit int) ([]LineWithPos, error) {
	var out []LineWithPos
	for i := len(s.chunks) - 1; i >= 0 && len(out) < limit; i-- {
		chunk := s.chunks[i]
		if s3LinePos(chunk.index, 0) >= start {
			continue
		}
		lines, err := s.chunkLines(chunk)
		if err != nil {
			return nil, err
		}
		for n := len(lines) - 1; n >= 0 && len(out) < limit; n-- {
			pos := s3LinePos(chunk.index, n)
			if pos >= start {
				continue
			}
			if entry, ok := parseValidLogLine(lines[n]); ok {
				out = append(out, LineWithPos{content: entry, startPos: pos, endPos: pos + 1})
			}
		}
	}
	// collected newest first; return oldest first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (s *s3LogSeries) readNewer(start int64, limit int) ([]LineWithPos, error) {
	var out []LineWithPos
	for _, chunk := range s.chunks {
		if len(out) >= limit {
			break
		}
		if s3LinePos(chunk.index+1, 0) <= start {
			continue
		}
		lines, err := s.chunkLines(chunk)
		if err != nil {
			return nil, err
		}
		for n := 0; n < len(lines) && len(out) < limit; n++ {
			pos := s3LinePos(chunk.index, n)
			if pos < start {
				continue
			}
			if entry, ok := parseValidLogLine(lines[n]); ok {
				out = append(out, LineWithPos{content: entry, startPos: pos, endPos: pos + 1})
			}
		}
	}
	return out, nil
}
