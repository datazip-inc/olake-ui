package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// LogEntry represents a log entry
type LogEntry struct {
	Level   string          `json:"level"`
	Time    time.Time       `json:"time"`
	Message json.RawMessage `json:"message"` // store raw JSON
}

type LineWithPos struct {
	content  string
	startPos int64 // byte position where this line starts
}

// isValidLogLine checks if a line is a valid, non-debug log entry
func isValidLogLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	var logEntry LogEntry
	if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
		return false
	}

	if logEntry.Level == "debug" {
		return false
	}

	return true
}

// ReadLinesBackward reads up to `limit` complete VALID log lines from file backwards starting at startOffset.
// Filters out empty lines, invalid JSON, and debug-level logs DURING reading.
// startOffset is treated as exclusive - we read lines that END BEFORE startOffset.
// Returns: valid lines (oldest->newest), newOffset (byte position before first returned line), hasMore, error.
func ReadLinesBackward(f *os.File, startOffset int64, limit int, fileSize int64) ([]string, int64, bool, error) {
	if limit <= 0 {
		return nil, 0, false, fmt.Errorf("limit must be greater than 0")
	}

	// startOffset beyond file size, clamp it to file size
	startOffset = min(startOffset, fileSize)

	// startOffset at beginning or negative, return empty result
	if startOffset <= 0 {
		return []string{}, 0, false, nil
	}

	offset := startOffset

	// 'tail' holds the partial line fragment at the start of a chunk.
	// This will be prepended to the NEXT chunk (which comes chronologically BEFORE this one).
	var tail []byte

	// valid lines we collected (newest first)
	foundLines := make([]LineWithPos, 0, limit)

	// read lines backwards until we have enough VALID lines or reach the beginning of the file
	for offset > 0 && len(foundLines) < limit {
		toRead := min(offset, int64(constants.LogReadChunkSize))
		readPos := offset - toRead

		chunk := make([]byte, toRead)
		n, rerr := f.ReadAt(chunk, readPos)

		// io.EOF is expected when reading near file boundaries
		if rerr != nil && rerr != io.EOF {
			return nil, 0, false, rerr
		}

		if int64(n) != toRead {
			chunk = chunk[:n]
		}

		data := make([]byte, 0, len(chunk)+len(tail))
		data = append(data, chunk...)
		data = append(data, tail...)

		// Extract lines from Right to Left
		for len(foundLines) < limit {
			// Find the last newline in the buffer
			lastNL := bytes.LastIndexByte(data, '\n')

			if lastNL == -1 {
				// No more newlines. The whole buffer is a partial line.
				// Save it as 'tail' for the next iteration.
				tail = data
				break
			}

			// The text AFTER the newline is a complete log line
			lineBytes := data[lastNL+1:]
			lineContent := string(lineBytes)

			// Calculate absolute file position of this line
			// readPos (start of chunk) + lastNL (relative index) + 1 (char after \n)
			linePos := readPos + int64(lastNL) + 1

			if isValidLogLine(lineContent) {
				foundLines = append(foundLines, LineWithPos{
					content:  lineContent,
					startPos: linePos,
				})
			}

			// CROP the buffer: remove the line we just processed
			data = data[:lastNL]
		}

		offset = readPos
		if offset == 0 {
			// Process the first line of the file if it's in the tail
			if len(tail) > 0 && len(foundLines) < limit {
				lineContent := string(tail)
				if isValidLogLine(lineContent) {
					foundLines = append(foundLines, LineWithPos{
						content:  lineContent,
						startPos: 0, // First line starts at position 0
					})
				}
			}
			break
		}
	}

	// no valid lines found
	if len(foundLines) == 0 {
		return []string{}, 0, false, nil
	}

	// Extract just the line content for return
	lines := make([]string, len(foundLines))
	for i, line := range foundLines {
		lines[len(foundLines)-1-i] = line.content // Reverse order
	}

	// The oldest line we returned is the last element in newestFirst
	newOffset := foundLines[len(foundLines)-1].startPos

	// hasMore is true only if we hit the limit with more file content remaining
	hasMore := newOffset > 0 && len(foundLines) == limit

	// If no more logs exist, return cursor at beginning (0)
	if !hasMore {
		newOffset = 0
	}

	return lines, newOffset, hasMore, nil
}

// ReadLinesForward reads up to `limit` complete VALID log lines from file forwards starting at startOffset.
// Filters out empty lines, invalid JSON, and debug-level logs DURING reading.
// startOffset is treated as inclusive - we start reading from exactly that position.
// Returns: valid lines (oldest->newest), newOffset (byte position after last returned line), hasMore, error.
func ReadLinesForward(f *os.File, startOffset int64, limit int, fileSize int64) ([]string, int64, bool, error) {
	if limit <= 0 {
		return nil, 0, false, fmt.Errorf("limit must be greater than 0")
	}

	// Ensure startOffset is at least 0
	startOffset = max(startOffset, 0)

	// If already at or past EOF, nothing to read
	if startOffset >= fileSize {
		return []string{}, fileSize, false, nil
	}

	// Seek to the startOffset position in the file before beginning to read lines
	if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
		return nil, 0, false, err
	}

	reader := bufio.NewReader(f)

	lines := make([]string, 0, limit)
	currentOffset := startOffset

	for len(lines) < limit {
		lineBytes, rerr := reader.ReadBytes('\n')

		if len(lineBytes) > 0 {
			// Update offset by bytes read
			currentOffset += int64(len(lineBytes))

			// Remove trailing newline and check if valid
			line := strings.TrimRight(string(lineBytes), "\r\n")
			if isValidLogLine(line) {
				lines = append(lines, line)
			}
		}

		if rerr != nil {
			// ReadBytes may return data and io.EOF together
			// so treat EOF as normal end-of-file and stop reading, return other errors
			if rerr == io.EOF {
				break
			}
			return nil, 0, false, rerr
		}
	}

	// hasMore is true only if we hit the limit with more file content remaining
	hasMore := currentOffset < fileSize && len(lines) == limit

	// If no more logs exist, return cursor at end (fileSize)
	if !hasMore {
		currentOffset = fileSize
	}

	return lines, currentOffset, hasMore, nil
}

// ReadLogs reads logs from the given mainLogDir and returns structured log entries.
// Direction can be "older" or "newer". If cursor < 0, it tails from the end of the file.
// Returns a TaskLogsResponse-like struct: oldest->newest logs plus cursors and hasMore flags.
func ReadLogs(mainLogDir string, cursor int64, limit int, direction string) (*dto.TaskLogsResponse, error) {
	// Check if mainLogDir exists
	if _, err := os.Stat(mainLogDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("logs directory not found: %s: %s", mainLogDir, err)
	}

	// Resolve and validate logs/sync_* directory
	logsDir, syncFolderName, err := GetAndValidateSyncDir(mainLogDir)
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(logsDir, syncFolderName)
	logPath := filepath.Join(logDir, "olake.log")

	logFile, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %s: %s", logPath, err)
	}
	defer logFile.Close()

	stat, err := logFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat log file: %s", err)
	}
	fileSize := stat.Size()

	// Normalize limit
	if limit <= 0 {
		limit = constants.DefaultLogsLimit
	}

	// Clamp cursor to file size
	if cursor > fileSize {
		cursor = fileSize
	}

	// Normalize direction
	dir := strings.ToLower(strings.TrimSpace(direction))
	if dir != "newer" {
		dir = constants.DefaultLogsDirection
	}

	// Initial tail: cursor < 0 means "from end of file"
	isTail := cursor < 0

	response := &dto.TaskLogsResponse{}

	// Parse validated lines into response format
	// Lines are already filtered (no empty, no invalid JSON, no debug) by ReadLines functions
	parseLines := func(lines []string) []map[string]interface{} {
		batch := make([]map[string]interface{}, 0, len(lines))
		for _, line := range lines {
			var logEntry LogEntry

			if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
				continue
			}

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

	// Tail or "older" from a cursor: walk backwards
	if isTail || dir == "older" {
		if isTail {
			cursor = fileSize
		}

		lines, newOffset, more, rerr := ReadLinesBackward(logFile, cursor, limit, fileSize)
		if rerr != nil {
			return nil, rerr
		}

		response.Logs = parseLines(lines)

		// olderCursor points to the position BEFORE the oldest log we're returning
		response.OlderCursor = newOffset
		response.NewerCursor = cursor
		response.HasMoreOlder = more
		response.HasMoreNewer = response.NewerCursor < fileSize
	} else {
		// dir == "newer": walk forwards
		lines, newOffset, more, rerr := ReadLinesForward(logFile, cursor, limit, fileSize)
		if rerr != nil {
			return nil, rerr
		}

		response.Logs = parseLines(lines)

		// newerCursor points to the position AFTER the newest log we have
		response.NewerCursor = newOffset
		response.OlderCursor = cursor
		response.HasMoreNewer = more
		response.HasMoreOlder = response.OlderCursor > 0
	}

	return response, nil
}
