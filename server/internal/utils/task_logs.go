package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// workerLogFileName is written by the olake worker next to the connector's sync_* folders.
const workerLogFileName = "worker.log"

// taskLogEntry is one log line with the fields the task logs page needs.
// Unlike LogEntry it keeps the worker's category and tip fields.
type taskLogEntry struct {
	Level    string          `json:"level"`
	Time     time.Time       `json:"time"`
	Message  json.RawMessage `json:"message"`
	Category string          `json:"category"`
	Tip      bool            `json:"tip"`
}

// logFile is one source file taking part in a task logs read.
type logFile struct {
	source string
	path   string
	file   *os.File
	size   int64
}

// taskLogLine is a parsed line plus where it came from, used while merging.
type taskLogLine struct {
	entry  taskLogEntry
	source string
	order  int // position of the source in the merge, breaks timestamp ties
	pos    LineWithPos
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
	// clients from before the merged view send a plain olake.log byte offset,
	// with -1 meaning tail; base64 JSON never parses as a number
	if offset, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if offset < 0 {
			return taskLogCursor{}, nil
		}
		return taskLogCursor{LogSourceSync: offset}, nil
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
	if err != nil {
		return nil, err
	}
	defer closeLogFiles(files)

	// start offset per file; tail (no cursor) starts at the end of every file
	start := taskLogCursor{}
	for _, f := range files {
		offset, ok := cursor[f.source]
		if !ok {
			offset = f.size
		}
		start[f.source] = min(max(offset, 0), f.size)
	}

	var candidates []taskLogLine
	// a file that returned fewer lines than asked has nothing further in this direction
	exhausted := map[string]bool{}
	for i, f := range files {
		var lines []LineWithPos
		if older {
			lines, err = readLinesBackwardWithPos(f.file, start[f.source], limit, f.size)
		} else {
			lines, _, err = readLinesForwardWithPos(f.file, start[f.source], limit, f.size)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read %s logs: %s", f.source, err)
		}
		exhausted[f.source] = len(lines) < limit
		for _, line := range lines {
			var entry taskLogEntry
			if err := json.Unmarshal([]byte(line.content), &entry); err != nil {
				continue
			}
			candidates = append(candidates, taskLogLine{entry: entry, source: f.source, order: i, pos: line})
		}
	}

	// each file is already in order; a stable sort merges them by time
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if !a.entry.Time.Equal(b.entry.Time) {
			return a.entry.Time.Before(b.entry.Time)
		}
		return a.order < b.order
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

	// the next offset per file is set by the furthest line returned from it
	next := taskLogCursor{}
	for source, offset := range start {
		next[source] = offset
	}
	// lines read from a file but cut by the limit
	unreturned := map[string]int{}
	for i := range candidates {
		unreturned[candidates[i].source]++
	}
	for i := range selected {
		line := &selected[i]
		unreturned[line.source]--
		if older {
			next[line.source] = min(next[line.source], line.pos.startPos)
		} else {
			next[line.source] = max(next[line.source], line.pos.endPos)
		}
	}
	// every line read from an exhausted file was returned: jump past what is
	// left (only debug/invalid lines) so paging stops
	for _, f := range files {
		if exhausted[f.source] && unreturned[f.source] == 0 {
			if older {
				next[f.source] = 0
			} else {
				next[f.source] = f.size
			}
		}
	}

	response := &dto.JobTaskLogsResponse{Logs: make([]map[string]interface{}, 0, len(selected))}
	for i := range selected {
		response.Logs = append(response.Logs, selected[i].toMap())
	}

	moreBeyond := false
	for _, f := range files {
		if older && next[f.source] > 0 || !older && next[f.source] < f.size {
			moreBeyond = true
		}
	}
	moreBehind := false
	for _, f := range files {
		if older && start[f.source] < f.size || !older && start[f.source] > 0 {
			moreBehind = true
		}
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
			paths = append(paths, logFile{source: LogSourceSync, path: filepath.Join(logsDir, syncFolder, "olake.log")})
		}
	}
	if source == LogSourceAll || source == LogSourceWorker {
		paths = append(paths, logFile{source: LogSourceWorker, path: filepath.Join(logsDir, workerLogFileName)})
	}

	files := make([]logFile, 0, len(paths))
	for _, f := range paths {
		file, err := os.Open(f.path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			closeLogFiles(files)
			return nil, fmt.Errorf("failed to read log file: %s: %s", f.path, err)
		}
		stat, err := file.Stat()
		if err != nil {
			file.Close()
			closeLogFiles(files)
			return nil, fmt.Errorf("failed to stat log file: %s", err)
		}
		f.file, f.size = file, stat.Size()
		files = append(files, f)
	}
	return files, nil
}

func closeLogFiles(files []logFile) {
	for _, f := range files {
		f.file.Close()
	}
}

func (l *taskLogLine) toMap() map[string]interface{} {
	out := map[string]interface{}{
		"level":   l.entry.Level,
		"time":    l.entry.Time.UTC().Format(time.RFC3339),
		"message": logMessageString(l.entry.Message),
		"source":  l.source,
	}
	if l.entry.Category != "" {
		out["category"] = l.entry.Category
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
