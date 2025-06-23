package shared

// Command represents the operation type for Kubernetes Jobs
type Command string

const (
	Discover Command = "discover"
	Check    Command = "check"
	Sync     Command = "sync"
)

// ActivityParams contains parameters for Kubernetes Job activities
// Copy from server/internal/temporal/types.go but remove Docker dependency
type ActivityParams struct {
	SourceType   string
	Version      string
	Config       string
	SourceID     int
	Command      Command
	DestConfig   string
	DestID       int
	WorkflowID   string
	StreamConfig string
	Flag         string
}

// SyncParams contains parameters for sync activities
type SyncParams struct {
	JobID      int
	WorkflowID string
}

// JobConfig represents configuration files for a Kubernetes Job
type JobConfig struct {
	Name string
	Data string
}
