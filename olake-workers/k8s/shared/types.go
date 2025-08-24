package shared

// Command represents the operation type for Kubernetes Jobs
type Command string

const (
	Discover Command = "discover"
	Check    Command = "check"
	Sync     Command = "sync"
)

// ActivityParams contains parameters for Kubernetes Job activities
// Must match server/internal/temporal/types.go exactly (no JSON tags)
type ActivityParams struct {
	SourceType    string
	Version       string
	Config        string
	SourceID      int
	Command       Command
	DestConfig    string
	DestID        int
	WorkflowID    string
	StreamsConfig string
	JobID         int
	Flag          string
}

// SyncParams contains parameters for sync activities
// Must match server/internal/temporal/types.go exactly (no JSON tags)
type SyncParams struct {
	JobID      int
	WorkflowID string
}

// JobConfig represents configuration files for a Kubernetes Job
type JobConfig struct {
	Name string `json:"name"`
	Data string `json:"data"`
}
