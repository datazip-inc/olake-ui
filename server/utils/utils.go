package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/oklog/ulid"
	"github.com/robfig/cron"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
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
	ctx.Data["json"] = models.JSONResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	_ = ctx.ServeJSON()
}

func SuccessResponse(ctx *web.Controller, data interface{}) {
	RespondJSON(ctx, http.StatusOK, true, "success", data)
}

func ErrorResponse(ctx *web.Controller, status int, message string) {
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
		logs.Critical(err)
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
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
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
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}
	return nil
}

// ParseJSONFile parses a JSON file into a map
func ParseJSONFile(filePath string) (map[string]interface{}, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(fileData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %v", filePath, err)
	}

	return result, nil
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

func CleanOldLogs(logDir string, retentionPeriod int) {
	logs.Info("Running log cleaner...")
	cutoff := time.Now().AddDate(0, 0, -retentionPeriod)

	// check if old logs are present
	shouldDelete := func(path string, cutoff time.Time) bool {
		entries, _ := os.ReadDir(path)
		if len(entries) == 0 {
			return true
		}

		var foundOldLog bool
		_ = filepath.Walk(path, func(filePath string, info os.FileInfo, _ error) error {
			if info == nil || info.IsDir() {
				return nil
			}
			if (strings.HasSuffix(filePath, ".log") || strings.HasSuffix(filePath, ".log.gz")) &&
				info.ModTime().Before(cutoff) {
				foundOldLog = true
				return filepath.SkipDir
			}
			return nil
		})
		return foundOldLog
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		logs.Error("failed to read log dir: %v", err)
		return
	}
	// delete dir if old logs are found or is empty
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "telemetry" {
			continue
		}
		dirPath := filepath.Join(logDir, entry.Name())
		if toDelete := shouldDelete(dirPath, cutoff); toDelete {
			logs.Info("Deleting folder: %s", dirPath)
			_ = os.RemoveAll(dirPath)
		}
	}
}

// starts a log cleaner that removes old logs from the specified directory based on the retention period
func InitLogCleaner(logDir string, retentionPeriod int) {
	logs.Info("Log cleaner started...")
	CleanOldLogs(logDir, retentionPeriod) // catchup missed cycles if any
	c := cron.New()
	err := c.AddFunc("@midnight", func() {
		CleanOldLogs(logDir, retentionPeriod)
	})
	if err != nil {
		logs.Error("Failed to start log cleaner: %v", err)
		return
	}
	c.Start()
}

// GetRetentionPeriod returns the retention period for logs
func GetLogRetentionPeriod() int {
	if val := os.Getenv("LOG_RETENTION_PERIOD"); val != "" {
		if retentionPeriod, err := strconv.Atoi(val); err == nil && retentionPeriod > 0 {
			return retentionPeriod
		}
	}
	return constants.DefaultLogRetentionPeriod
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

// ReadLogs reads logs from the given mainLogDir and returns structured log entries.
func ReadLogs(mainLogDir string) ([]map[string]interface{}, error) {
	// Check if mainLogDir exists
	if _, err := os.Stat(mainLogDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("logs directory not found: %s", mainLogDir)
	}

	// Logs directory
	logsDir := filepath.Join(mainLogDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("logs directory not found: %s", logsDir)
	}

	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("logs directory empty in: %s", logsDir)
	}

	logDir := filepath.Join(logsDir, files[0].Name())
	logPath := filepath.Join(logDir, "olake.log")
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %s", logPath)
	}

	var parsedLogs []map[string]interface{}
	lines := strings.Split(string(logContent), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var logEntry LogEntry
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		if logEntry.Level == "debug" {
			continue
		}

		// Convert Message to string safely
		var messageStr string
		var tmp interface{}
		if err := json.Unmarshal(logEntry.Message, &tmp); err == nil {
			switch v := tmp.(type) {
			case string:
				messageStr = v // plain string
			default:
				msgBytes, _ := json.Marshal(v) // object/array
				messageStr = string(msgBytes)
			}
		} else {
			// fallback: raw bytes as string
			messageStr = string(logEntry.Message)
		}

		parsedLogs = append(parsedLogs, map[string]interface{}{
			"level":   logEntry.Level,
			"time":    logEntry.Time.UTC().Format(time.RFC3339),
			"message": messageStr,
		})
	}

	return parsedLogs, nil
}
