package constants

const (
	// LogReadChunkSize controls log file read chunks (64KB)
	LogReadChunkSize = 64 * 1024

	// DefaultLogsLimit is the fallback log count per request
	DefaultLogsLimit = 1000

	// DefaultLogsCursor represents tail-from-end behavior
	DefaultLogsCursor int64 = -1
)

