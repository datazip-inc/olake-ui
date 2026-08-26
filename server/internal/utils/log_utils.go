package utils

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// LogEntry represents a log entry
type LogEntry struct {
	Level   string          `json:"level"`
	Time    time.Time       `json:"time"`
	Message json.RawMessage `json:"message"` // store raw JSON
}

func ReadLogs(ctx context.Context, workflowID string, cursor int64, limit int, direction string) (*dto.TaskLogsResponse, error) {
	mode := strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode))
	switch mode {
	case constants.StorageModeS3:
		return readLogsFromS3(ctx, workflowID, cursor, limit, direction)
	default:
		return readLogsFromNFS(workflowID, cursor, limit, direction)
	}
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

func parseLines(lines []string) []map[string]interface{} {
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
