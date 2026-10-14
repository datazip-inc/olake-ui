package utils

import (
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
)

// Log sources shown on the task logs page.
const (
	LogSourceAll    = "all"
	LogSourceSync   = "sync"
	LogSourceWorker = "worker"
)

// taskLogEntry is one log line with the fields the task logs page needs.
// Unlike LogEntry it keeps the worker's tip flag.
type taskLogEntry struct {
	Level   string          `json:"level"`
	Time    time.Time       `json:"time"`
	Message json.RawMessage `json:"message"`
	Tip     bool            `json:"tip"`
}

// logFile is one source file taking part in a task logs read, with its
// paging state for the current request.
type logFile struct {
	source string
	path   string
	file   *os.File
	size   int64

	start      int64 // where this read starts in the file
	next       int64 // where the following read continues
	exhausted  bool  // fewer lines than asked: nothing further in this direction
	unreturned int   // lines read but cut by the limit
}

// taskLogLine is a parsed line plus where it came from, used while merging.
type taskLogLine struct {
	entry taskLogEntry
	file  int // index into files; also the tie-break order (sync first)
	pos   LineWithPos
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

// taskLogCursor holds one byte offset per source file. It travels to the
// client as an opaque base64 string; an empty string means "tail from the end".
type taskLogCursor map[string]int64

func decodeTaskLogCursor(raw string) (taskLogCursor, error) {
	if raw == "" {
		return taskLogCursor{}, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid log cursor: %s", err)
	}
	var cursor taskLogCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid log cursor: %s", err)
	}
	return cursor, nil
}

func (c taskLogCursor) encode() string {
	data, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(data)
}

// ReadTaskLogs reads a run's logs from the connector's olake.log and/or the
// worker's worker.log. With source "all" both files are merged by timestamp.
// Pagination uses one byte offset per file, carried in an opaque cursor.
// A missing file counts as empty: old workers write no worker.log, and the
// connector creates its sync_* folder only once it starts.
func ReadTaskLogs(baseDir, source, rawCursor string, limit int, direction string) (*dto.JobTaskLogsResponse, error) {
	cursor, err := decodeTaskLogCursor(rawCursor)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = constants.DefaultLogsLimit
	}
	older := strings.ToLower(strings.TrimSpace(direction)) != "newer"

	files, err := openTaskLogFiles(baseDir, NormalizeLogSource(source))
	defer closeLogFiles(files)
	if err != nil {
		return nil, err
	}

	var candidates []taskLogLine
	for i := range files {
		f := &files[i]
		offset, ok := cursor[f.source]
		if !ok {
			// no cursor: tail from the end. A file missing from a non-empty
			// cursor was created after it was issued, so all of it is newer.
			offset = f.size
			if len(cursor) > 0 {
				offset = 0
			}
		}
		f.start = min(max(offset, 0), f.size)
		f.next = f.start

		var lines []LineWithPos
		if older {
			lines, err = readLinesBackwardWithPos(f.file, f.start, limit, f.size)
		} else {
			lines, _, err = readLinesForwardWithPos(f.file, f.start, limit, f.size)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read %s logs: %s", f.source, err)
		}
		f.exhausted = len(lines) < limit
		for _, line := range lines {
			var entry taskLogEntry
			if err := json.Unmarshal([]byte(line.content), &entry); err != nil {
				continue
			}
			candidates = append(candidates, taskLogLine{entry: entry, file: i, pos: line})
			f.unreturned++
		}
	}

	// each file is already in order; a stable sort merges them by time
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if !a.entry.Time.Equal(b.entry.Time) {
			return a.entry.Time.Before(b.entry.Time)
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

	// the next offset of a file is set by the furthest line returned from it
	response := &dto.JobTaskLogsResponse{Logs: make([]map[string]interface{}, 0, len(selected))}
	for i := range selected {
		line := &selected[i]
		f := &files[line.file]
		f.unreturned--
		if older {
			f.next = min(f.next, line.pos.startPos)
		} else {
			f.next = max(f.next, line.pos.endPos)
		}
		response.Logs = append(response.Logs, line.toMap(f.source))
	}

	start, next := taskLogCursor{}, taskLogCursor{}
	moreBeyond, moreBehind := false, false
	for i := range files {
		f := &files[i]
		// every line read from an exhausted file was returned: jump past what
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

	var paths []logFile
	if source == LogSourceAll || source == LogSourceSync {
		if _, syncFolder, err := GetAndValidateSyncDir(baseDir); err == nil {
			paths = append(paths, logFile{source: LogSourceSync, path: filepath.Join(logsDir, syncFolder, constants.SyncLogFileName)})
		}
	}
	if source == LogSourceAll || source == LogSourceWorker {
		paths = append(paths, logFile{source: LogSourceWorker, path: filepath.Join(logsDir, constants.WorkerLogFileName)})
	}

	files := make([]logFile, 0, len(paths))
	for _, f := range paths {
		file, err := os.Open(f.path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return files, fmt.Errorf("failed to read log file: %s: %s", f.path, err)
		}
		// append before Stat so the file is closed even if Stat fails
		f.file = file
		files = append(files, f)
		stat, err := file.Stat()
		if err != nil {
			return files, fmt.Errorf("failed to stat log file: %s", err)
		}
		files[len(files)-1].size = stat.Size()
	}
	return files, nil
}

func closeLogFiles(files []logFile) {
	for i := range files {
		files[i].file.Close()
	}
}

func (l *taskLogLine) toMap(source string) map[string]interface{} {
	out := map[string]interface{}{
		"level":   l.entry.Level,
		"time":    l.entry.Time.UTC().Format(time.RFC3339),
		"message": logMessageString(l.entry.Message),
		"source":  source,
	}
	if l.entry.Tip {
		out["tip"] = true
	}
	return out
}

// logMessageString returns a log message as text: strings as-is, any other
// JSON value re-marshalled.
func logMessageString(raw json.RawMessage) string {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	if str, ok := value.(string); ok {
		return str
	}
	msgBytes, err := json.Marshal(value)
	if err != nil {
		return string(raw)
	}
	return string(msgBytes)
}
