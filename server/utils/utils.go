package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/oklog/ulid"

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

const readChunkSize = 128 * 1024 // read files in chunks of 128KB

// ReadLinesBackward reads up to `limit` complete lines from file backwards starting at startOffset.
// Returns: lines (oldest->newest), newOffset (byte position before first returned line), hasMore, error.
func ReadLinesBackward(f *os.File, startOffset int64, limit int) ([]string, int64, bool, error) {
	if limit <= 0 {
		return nil, 0, false, fmt.Errorf("limit must be greater than 0")
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, 0, false, err
	}

	fileSize := fi.Size()

	// startOffset beyond file size, clamp it to file size
	startOffset = min(startOffset, fileSize)

	// startOffset at beginning or negative, return empty result
	if startOffset <= 0 {
		return nil, 0, false, nil
	}

	offset := startOffset

	// accumulator to store the lines we read backwards
	var accum []byte

	// lines we read backwards
	newestFirst := make([]string, 0, limit)

	// read lines backwards until we have enough lines or reach the beginning of the file
	for offset > 0 && len(newestFirst) < limit {
		toRead := min(offset, int64(readChunkSize))
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

		// prepend this chunk to accum
		tmp := make([]byte, 0, len(chunk)+len(accum))
		tmp = append(tmp, chunk...)
		tmp = append(tmp, accum...)
		accum = tmp

		parts := bytes.Split(accum, []byte{'\n'})

		// last part may be incomplete line (no trailing \n) unless at file start
		fullCount := len(parts) - 1
		if readPos == 0 {
			// at file start, all parts are complete (even without trailing \n)
			fullCount = len(parts)
		}

		// Collect complete lines from newest to oldest
		added := 0
		for i := len(parts) - 1; i >= 0 && added < fullCount && len(newestFirst) < limit; i-- {
			newestFirst = append(newestFirst, string(parts[i]))
			added++
		}

		offset = readPos
		if offset == 0 {
			break
		}
	}

	// no complete lines found in accumulator
	if len(newestFirst) == 0 {
		return []string{}, offset, offset > 0, nil
	}

	// find where oldest returned line starts
	parts := bytes.Split(accum, []byte{'\n'})
	returnedCount := len(newestFirst)
	lastIdx := len(parts) - 1
	firstReturnedIdx := lastIdx - (returnedCount - 1)

	// count bytes before the first returned line
	bytesBefore := 0
	for i := 0; i < firstReturnedIdx; i++ {
		bytesBefore += len(parts[i]) + 1 // +1 for '\n'
	}
	newOffset := offset + int64(bytesBefore)

	// Reverse to return oldest->newest order
	slices.Reverse(newestFirst)

	hasMore := newOffset > 0
	return newestFirst, newOffset, hasMore, nil

}

// ReadLogs reads logs from the given mainLogDir and returns structured log entries.
// Returns: lines (oldest->newest), newOffset (byte position before first returned line), hasMore, error.
func ReadLogs(mainLogDir string, cursor int64, limit int) ([]map[string]interface{}, int64, bool, error) {
	// Check if mainLogDir exists
	if _, err := os.Stat(mainLogDir); os.IsNotExist(err) {
		return nil, 0, false, fmt.Errorf("logs directory not found: %s", mainLogDir)
	}

	// Logs directory
	logsDir := filepath.Join(mainLogDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil, 0, false, fmt.Errorf("logs directory not found: %s", logsDir)
	}

	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		return nil, 0, false, fmt.Errorf("logs directory empty in: %s", logsDir)
	}

	logDir := filepath.Join(logsDir, files[0].Name())
	logPath := filepath.Join(logDir, "olake.log")

	f, err := os.Open(logPath)
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to read log file: %s", logPath)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to stat log file: %w", err)
	}
	fileSize := stat.Size()

	// Normalize limit
	if limit <= 0 {
		limit = 1000
	}

	// Normalize cursor: negative means from the EOF
	if cursor < 0 {
		cursor = fileSize
	}

	// Clamp cursor to file size
	if cursor > fileSize {
		cursor = fileSize
	}

	// If cursor is 0, we've read everything
	if cursor == 0 {
		return []map[string]interface{}{}, 0, false, nil
	}

	var parsedLogs []map[string]interface{}
	currentOffset := cursor
	hasMore := false
	for len(parsedLogs) < limit {
		// fetch a chunk: we request a bit more lines (multiplier) to account for filtered lines
		fetchLines := max((limit-len(parsedLogs))*2, 1)
		lines, newOffset, more, rerr := ReadLinesBackward(f, currentOffset, fetchLines)
		if rerr != nil {
			return nil, 0, false, rerr
		}
		if len(lines) == 0 {
			// no more lines to read
			hasMore = false
			currentOffset = newOffset
			break
		}

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			var logEntry LogEntry
			if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
				continue
			}

			if logEntry.Level == "debug" {
				continue
			}

			var messageStr string
			var tmp interface{}
			if err := json.Unmarshal(logEntry.Message, &tmp); err == nil {
				switch v := tmp.(type) {
				case string:
					messageStr = v
				default:
					msgBytes, _ := json.Marshal(v)
					messageStr = string(msgBytes)
				}
			} else {
				messageStr = string(logEntry.Message)
			}

			parsedLogs = append(parsedLogs, map[string]interface{}{
				"level":   logEntry.Level,
				"time":    logEntry.Time.UTC().Format(time.RFC3339),
				"message": messageStr,
			})
			if len(parsedLogs) >= limit {
				break
			}
		}
		currentOffset = newOffset
		hasMore = more

		if !more || currentOffset == 0 {
			break
		}
	}

	return parsedLogs, currentOffset, hasMore, nil
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
