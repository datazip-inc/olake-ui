package utils

import (
	"archive/tar"
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/beego/beego/v2/server/web"
	"github.com/oklog/ulid"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

func ToMapOfInterface(structure any) map[string]interface{} {
	if structure == nil {
		return nil
	}

	data, _ := json.Marshal(structure)

	var output map[string]interface{}
	_ = json.Unmarshal(data, &output)

	return output
}

func RespondJSON(ctx *web.Controller, status int, success bool, message string, data interface{}) {
	ctx.Ctx.Output.SetStatus(status)
	ctx.Data["json"] = dto.JSONResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	_ = ctx.ServeJSON()
}

func SuccessResponse(ctx *web.Controller, message string, data interface{}) {
	RespondJSON(ctx, http.StatusOK, true, message, data)
}

func ErrorResponse(ctx *web.Controller, status int, message string, err error) {
	if err != nil {
		logger.Errorf("error in request %s: %s", ctx.Ctx.Input.URI(), err)
	}
	RespondJSON(ctx, status, false, message, nil)
}

func HandleJSONOK(w http.ResponseWriter, content interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(content)
}

// send a message as response
func HandleResponseMessage(w http.ResponseWriter, statusCode int, content interface{}, message string) {
	body := make(map[string]interface{})

	if content != nil {
		jsonbody, err := json.Marshal(content)
		if err != nil {
			HandleError(w, http.StatusInternalServerError, err)
		}

		if err = json.Unmarshal(jsonbody, &body); err != nil {
			HandleError(w, http.StatusInternalServerError, err)
		}
	}
	body["message"] = message

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

// send error as json response
func HandleErrorAsMessage(w http.ResponseWriter, statusCode int, err error) {
	body := make(map[string]string)
	body["error"] = err.Error()

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

// send error as direct text/string
func HandleError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, err)
}

// Handle errors and pass it to /error page
func HandleErrorJS(w http.ResponseWriter, r *http.Request, err error) {
	http.Redirect(w, r, fmt.Sprintf(`/error?msg=%q`, url.QueryEscape(err.Error())), http.StatusPermanentRedirect)
}

func ExistsInArray[T comparable](arr []T, value T) bool {
	for _, elem := range arr {
		if elem == value {
			return true
		}
	}

	return false
}

func ULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)

	t := time.Now()
	newUlid, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		logger.Fatal(err)
	}

	return newUlid.String()
}

func Ternary(cond bool, a, b any) any {
	if cond {
		return a
	}
	return b
}

// ExtractJobIDFromWorkflowID extracts the JobID from Temporal workflow IDs created by this system.
//
// Expected workflow ID shapes:
// - sync-<projectID>-<jobID>
// - sync-<projectID>-<jobID>-<suffix>
//
// projectID itself can contain '-', so we match the exact prefix and then parse the leading integer.
func ExtractJobIDFromWorkflowID(workflowID, projectID string) (int, bool) {
	prefix := "sync-" + projectID + "-"

	rest, ok := strings.CutPrefix(workflowID, prefix)
	if !ok || rest == "" {
		return 0, false
	}

	// Find the numeric prefix.
	i := 0
	for ; i < len(rest); i++ {
		if !unicode.IsDigit(rune(rest[i])) {
			break
		}
	}

	if i == 0 { // No leading digits
		return 0, false
	}

	id, err := strconv.Atoi(rest[:i])
	if err != nil {
		return 0, false
	}

	return id, true
}

// CreateDirectory creates a directory with the specified permissions if it doesn't exist
func CreateDirectory(dirPath string, perm os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %s", dirPath, err)
		}
	}
	return nil
}

// WriteFile writes data to a file, creating the directory if necessary
func WriteFile(filePath string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filePath)
	if err := CreateDirectory(dirPath, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, perm); err != nil {
		return fmt.Errorf("failed to write to file %s: %s", filePath, err)
	}
	return nil
}

// ToCron converts a frequency string to a cron expression
func ToCron(frequency string) string {
	parts := strings.Split(strings.ToLower(frequency), "-")
	if len(parts) != 2 {
		return frequency
	}

	valueStr, unit := parts[0], parts[1]
	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return frequency
	}

	switch unit {
	case "minutes":
		return fmt.Sprintf("*/%d * * * *", value) // Every N minutes
	case "hours":
		return fmt.Sprintf("0 */%d * * *", value) // Every N hours at minute 0
	case "days":
		return fmt.Sprintf("0 0 */%d * *", value) // Every N days at midnight
	case "weeks":
		// Every N weeks on Sunday (0), cron doesn't support "every N weeks" directly,
		// so simulate with day-of-week field (best-effort)
		return fmt.Sprintf("0 0 * * */%d", value)
	case "months":
		return fmt.Sprintf("0 0 1 */%d *", value) // Every N months on the 1st at midnight
	case "years":
		return fmt.Sprintf("0 0 1 1 */%d", value) // Every N years on the 1st of January at midnight
	default:
		return frequency
	}
}

// ExtractJSON extracts and returns the last valid JSON block from output
func ExtractJSON(output string) (map[string]interface{}, error) {
	outputStr := strings.TrimSpace(output)
	if outputStr == "" {
		return nil, fmt.Errorf("empty output")
	}

	lines := strings.Split(outputStr, "\n")

	// Find the last non-empty line with valid JSON
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		start := strings.Index(line, "{")
		end := strings.LastIndex(line, "}")
		if start != -1 && end != -1 && end > start {
			jsonPart := line[start : end+1]
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(jsonPart), &result); err != nil {
				continue // Skip invalid JSON
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no valid JSON block found in output")
}

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

// RetryWithBackoff retries a function with exponential backoff
func RetryWithBackoff(fn func() error, maxRetries int, initialDelay time.Duration) error {
	delay := initialDelay
	var errMsg error

	for retry := 0; retry < maxRetries; retry++ {
		if err := fn(); err != nil {
			errMsg = err
			if retry < maxRetries-1 {
				logger.Warnf("Retry attempt %d/%d failed: %s. Retrying in %v...", retry+1, maxRetries, err, delay)
				time.Sleep(delay)
				delay *= 2
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed after %d retries: %s", maxRetries, errMsg)
}

// GetAndValidateLogBaseDir returns the base directory path for log files
// based on the SHA256 hash of the filePath (workflow ID) and validates it exists
func GetAndValidateLogBaseDir(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))
	homeDir := constants.DefaultConfigDir
	baseDir := filepath.Join(homeDir, syncFolderName)

	// Verify directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return "", fmt.Errorf("logs directory not found: %s: %s", baseDir, err)
	}

	return baseDir, nil
}

// GetAndValidateSyncDir returns the logs directory and sync_* folder name under it
func GetAndValidateSyncDir(baseDir string) (string, string, error) {
	logsDir := filepath.Join(baseDir, "logs")

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to read logs directory: %s", err)
	}
	if len(entries) == 0 {
		return "", "", fmt.Errorf("no sync log folders found in: %s", logsDir)
	}

	for _, entry := range entries {
		// get the first directory that starts with "sync_"
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "sync_") {
			return logsDir, entry.Name(), nil
		}
	}

	return "", "", fmt.Errorf("no sync folder found in: %s", logsDir)
}

// addFileToArchive streams a file into the tar archive
func AddFileToArchive(tarWriter *tar.Writer, filePath, nameInArchive string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %s", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %s", filePath, err)
	}

	// tar header with file metadata
	header := &tar.Header{
		Name:    nameInArchive,
		Size:    fileInfo.Size(),
		Mode:    int64(fileInfo.Mode()),
		ModTime: fileInfo.ModTime(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %s", nameInArchive, err)
	}

	bytesWritten, err := io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf("failed to write file content for %s: %s", nameInArchive, err)
	}

	logger.Debugf("Added %s to archive (%d bytes)", nameInArchive, bytesWritten)

	return nil
}

// GetLogArchiveFilename generates the filename for the log archive download
func GetLogArchiveFilename(jobID int, filePath string) (string, error) {
	baseDir, err := GetAndValidateLogBaseDir(filePath)
	if err != nil {
		return "", err
	}

	_, syncFolderName, err := GetAndValidateSyncDir(baseDir)
	if err != nil {
		return "", err
	}

	syncTimestamp := strings.ReplaceAll(strings.TrimPrefix(syncFolderName, "sync_"), "_", "-")
	filename := fmt.Sprintf("job-%d-logs-%s.tar.gz", jobID, syncTimestamp)

	return filename, nil
}
