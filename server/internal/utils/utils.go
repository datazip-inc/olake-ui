package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/oklog/ulid"

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

func StringPtr(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
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

// WorkflowHash returns a deterministic hash string for a given workflowID.
func WorkflowHash(workflowID string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(workflowID)))
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

func ConvertMBToBytes(sizeMB int64) string {
	const bytesPerMB = 1024 * 1024
	sizeBytes := sizeMB * bytesPerMB
	return strconv.FormatInt(sizeBytes, 10)
}

// EqualJSON reports whether a and b represent the same JSON value.
// Invalid JSON on either side yields false.
func EqualJSON(a, b string) bool {
	if a == b {
		return true
	}
	var va, vb any
	if json.Unmarshal([]byte(a), &va) != nil || json.Unmarshal([]byte(b), &vb) != nil {
		return false
	}
	aBytes, err := json.Marshal(va)
	if err != nil {
		return false
	}
	bBytes, err := json.Marshal(vb)
	if err != nil {
		return false
	}
	return bytes.Equal(aBytes, bBytes)
}
