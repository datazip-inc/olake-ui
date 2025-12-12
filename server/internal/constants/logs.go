package constants

const (
	// LogReadChunkSize is the number of bytes read per chunk when scanning log files.
	LogReadChunkSize = 64 * 1024 // 64KB

	// DefaultLogsLimit is the number of log entries returned if no limit is provided.
	DefaultLogsLimit = 1000

	// DefaultLogsCursor indicates tailing from the end of the file (cursor < 0).
	DefaultLogsCursor int64 = -1

	// DefaultLogsDirection is the fallback pagination direction ("older" or "newer").
	DefaultLogsDirection = "older"
)
