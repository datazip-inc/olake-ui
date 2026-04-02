package utils

import (
	"archive/tar"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/oklog/ulid"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

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

func MarshalToString(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func SetIfNotEmpty(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	}
}

// NormalizeString converts a string to lowercase, trims leading and trailing spaces and replaces spaces with underscores
func NormalizeString(s string) string {
	words := strings.Fields(strings.ToLower(s))
	return strings.Join(words, "_")
}
