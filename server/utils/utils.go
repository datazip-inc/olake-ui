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

// ConnectionStatus represents the structure of the connection status JSON
type ConnectionStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// LogMessage represents the structure of the log message JSON
type LogMessage struct {
	ConnectionStatus *ConnectionStatus `json:"connectionStatus,omitempty"`
	Type             string            `json:"type,omitempty"`
	// Add other fields as needed
}

// ExtractAndParseLastLogMessage extracts the JSON from the last log line and parses it
func ExtractAndParseLastLogMessage(output []byte) (*LogMessage, error) {
	// Convert output to string and split into lines
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("empty output")
	}

	lines := strings.Split(outputStr, "\n")

	// Find the last non-empty line
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			lastLine = trimmed
			break
		}
	}

	if lastLine == "" {
		return nil, fmt.Errorf("no log lines found")
	}

	// Extract JSON part (everything after the first "{")
	start := strings.Index(lastLine, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON found in log line")
	}
	jsonStr := lastLine[start:]

	// Parse the JSON
	var msg LogMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &msg, nil
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
		return ""
	}

	valueStr, unit := parts[0], parts[1]
	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return ""
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
		return ""
	}
}

func CleanOldLogs(logDir string, retentionMinutes int) error {
	cutoff := time.Now().Add(-time.Minute * 3)

	shouldDelete := func(path string) (bool, error) {
		info, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		if !info.IsDir() || !info.ModTime().Before(cutoff) {
			return false, nil
		}

		// Check for any file ending in .log, .log.gz, or .gz
		found := false
		err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			name := d.Name()
			if strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".log.gz") {
				found = true
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return false, err
		}
		return found, nil
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "telemetry" {
			continue
		}

		dirPath := filepath.Join(logDir, entry.Name())
		ok, err := shouldDelete(dirPath)
		if err != nil {
			continue
		}
		if ok {
			fmt.Printf("Deleting folder %s\n", dirPath)
			if err := os.RemoveAll(dirPath); err != nil {
				continue
			}
		}
	}
	return nil
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
