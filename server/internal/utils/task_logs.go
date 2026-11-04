package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

// Log sources shown on the task logs page.
const (
	LogSourceAll    = "all"
	LogSourceSync   = "sync"
	LogSourceWorker = "worker"
)

// logFile is one log source taking part in a task logs read (a file on NFS,
// a series of chunks on S3), with its paging state for the current request.
// Positions are opaque to the merge: byte offsets on NFS, chunk and line
// numbers on S3. size is the end position: no line starts at or after it.
type logFile struct {
	source string
	size   int64

	// readOlder returns up to limit lines before start, readNewer up to limit
	// lines from start on; both oldest first.
	readOlder func(start int64, limit int) ([]LineWithPos, error)
	readNewer func(start int64, limit int) ([]LineWithPos, error)
	close     func()

	start      int64 // where this read starts in the source
	next       int64 // where the following read continues
	exhausted  bool  // fewer lines than asked: nothing further in this direction
	unreturned int   // lines read but cut by the limit
}

// taskLogLine is a parsed line plus where it came from, used while merging.
type taskLogLine struct {
	file int // index into files; also the tie-break order (sync first)
	line LineWithPos
}

// NormalizeLogSource returns a known log source, defaulting to "all".
func NormalizeLogSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case LogSourceSync:
		return LogSourceSync
	case LogSourceWorker:
		return LogSourceWorker
	default:
		return LogSourceAll
	}
}

// taskLogCursor holds one position per log source. It travels to the
// client as an opaque base64 string; an empty string means "tail from the end".
type taskLogCursor map[string]int64

func decodeTaskLogCursor(raw string) (taskLogCursor, error) {
	if raw == "" {
		return taskLogCursor{}, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", constants.ErrInvalidLogCursor, err)
	}
	var cursor taskLogCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("%w: %s", constants.ErrInvalidLogCursor, err)
	}
	return cursor, nil
}

func (c taskLogCursor) encode() string {
	data, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(data)
}

// ReadTaskLogs reads a page of a run's logs for the task logs page, from the
// storage the logs live in. The cursor is opaque to the client.
func ReadTaskLogs(ctx context.Context, baseDir, source, rawCursor string, limit int, direction string) (*dto.JobTaskLogsResponse, error) {
	cursor, err := decodeTaskLogCursor(rawCursor)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = constants.DefaultLogsLimit
	}
	source = NormalizeLogSource(source)

	var files []logFile
	if storage.Mode() == constants.StorageModeS3 {
		files, err = openS3TaskLogs(ctx, baseDir, source)
	} else {
		files, err = openTaskLogFiles(baseDir, source)
	}
	defer closeLogFiles(files)
	if err != nil {
		return nil, err
	}
	return mergeTaskLogs(files, cursor, limit, strings.ToLower(strings.TrimSpace(direction)) != "newer")
}

// mergeTaskLogs reads a page from the connector's and/or the worker's logs.
// With source "all" both are merged by timestamp. Pagination uses one
// position per source, carried in the cursor.
// A missing source counts as empty: old workers write no worker logs, and the
// connector creates its sync_* folder only once it starts.
func mergeTaskLogs(files []logFile, cursor taskLogCursor, limit int, older bool) (*dto.JobTaskLogsResponse, error) {
	var candidates []taskLogLine
	for i := range files {
		f := &files[i]
		offset, ok := cursor[f.source]
		if !ok {
			// no cursor: tail from the end. A source missing from a non-empty
			// cursor was created after it was issued, so all of it is newer.
			offset = f.size
			if len(cursor) > 0 {
				offset = 0
			}
		}
		f.start = min(max(offset, 0), f.size)
		f.next = f.start

		read := f.readNewer
		if older {
			read = f.readOlder
		}
		lines, err := read(f.start, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s logs: %s", f.source, err)
		}
		f.exhausted = len(lines) < limit
		for _, line := range lines {
			candidates = append(candidates, taskLogLine{file: i, line: line})
			f.unreturned++
		}
	}

	// each file is already in order; a stable sort merges them by time
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if !a.line.content.Time.Equal(b.line.content.Time) {
			return a.line.content.Time.Before(b.line.content.Time)
		}
		return a.file < b.file
	})

	// keep the lines nearest to the cursor
	selected := candidates
	if len(selected) > limit {
		if older {
			selected = selected[len(selected)-limit:]
		} else {
			selected = selected[:limit]
		}
	}

	// the next position of a source is set by the furthest line returned from it
	response := &dto.JobTaskLogsResponse{Logs: make([]map[string]interface{}, 0, len(selected))}
	for i := range selected {
		l := &selected[i]
		f := &files[l.file]
		f.unreturned--
		if older {
			f.next = min(f.next, l.line.startPos)
		} else {
			f.next = max(f.next, l.line.endPos)
		}
		response.Logs = append(response.Logs, l.toMap(f.source))
	}

	start, next := taskLogCursor{}, taskLogCursor{}
	moreBeyond, moreBehind := false, false
	for i := range files {
		f := &files[i]
		// every line read from an exhausted source was returned: jump past what
		// is left (only debug/invalid lines) so paging stops
		if f.exhausted && f.unreturned == 0 {
			f.next = 0
			if !older {
				f.next = f.size
			}
		}
		if older {
			moreBeyond = moreBeyond || f.next > 0
			moreBehind = moreBehind || f.start < f.size
		} else {
			moreBeyond = moreBeyond || f.next < f.size
			moreBehind = moreBehind || f.start > 0
		}
		start[f.source], next[f.source] = f.start, f.next
	}

	if older {
		response.OlderCursor, response.NewerCursor = next.encode(), start.encode()
		response.HasMoreOlder, response.HasMoreNewer = moreBeyond, moreBehind
	} else {
		response.OlderCursor, response.NewerCursor = start.encode(), next.encode()
		response.HasMoreOlder, response.HasMoreNewer = moreBehind, moreBeyond
	}
	return response, nil
}

// openTaskLogFiles opens the files for a source, in merge order (sync first).
func openTaskLogFiles(baseDir, source string) ([]logFile, error) {
	logsDir := filepath.Join(baseDir, "logs")

	type sourcePath struct{ source, path string }
	var paths []sourcePath
	if source == LogSourceAll || source == LogSourceSync {
		if _, syncFolder, err := GetAndValidateFilesystemSyncDir(baseDir); err == nil {
			paths = append(paths, sourcePath{LogSourceSync, filepath.Join(logsDir, syncFolder, constants.SyncLogFileName)})
		}
	}
	if source == LogSourceAll || source == LogSourceWorker {
		paths = append(paths, sourcePath{LogSourceWorker, filepath.Join(logsDir, constants.WorkerLogFileName)})
	}

	files := make([]logFile, 0, len(paths))
	for _, p := range paths {
		file, err := os.Open(p.path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return files, fmt.Errorf("failed to read log file: %s: %s", p.path, err)
		}
		// append before Stat so the file is closed even if Stat fails
		files = append(files, logFile{source: p.source, close: func() { file.Close() }})
		stat, err := file.Stat()
		if err != nil {
			return files, fmt.Errorf("failed to stat log file: %s", err)
		}
		size := stat.Size()
		f := &files[len(files)-1]
		f.size = size
		f.readOlder = func(start int64, limit int) ([]LineWithPos, error) {
			return readLinesBackwardWithPos(file, start, limit, size)
		}
		f.readNewer = func(start int64, limit int) ([]LineWithPos, error) {
			lines, _, err := readLinesForwardWithPos(file, start, limit, size)
			return lines, err
		}
	}
	return files, nil
}

func closeLogFiles(files []logFile) {
	for i := range files {
		if files[i].close != nil {
			files[i].close()
		}
	}
}

func (l *taskLogLine) toMap(source string) map[string]interface{} {
	entry := &l.line.content
	out := map[string]interface{}{
		"level":   entry.Level,
		"time":    entry.Time.UTC().Format(time.RFC3339),
		"message": logMessageString(entry.Message),
		"source":  source,
	}
	if entry.Tip {
		out["tip"] = true
	}
	return out
}
