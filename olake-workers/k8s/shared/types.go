package shared

import (
	"time"
)

// Command represents the operation type for Kubernetes Jobs
type Command string

const (
	Discover Command = "discover"
	Check    Command = "check"
	Sync     Command = "sync"
)

// String returns the string representation of the command
func (c Command) String() string {
	return string(c)
}

// ActivityParams contains parameters for Kubernetes Job activities
// Must match server/internal/temporal/types.go exactly (no JSON tags)
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
	JobID        int
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

// ActivityResult represents the result of an activity execution
type ActivityResult struct {
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// JobExecutionMetadata contains metadata about job execution
type JobExecutionMetadata struct {
	JobID             int       `json:"job_id"`
	WorkflowID        string    `json:"workflow_id"`
	KubernetesJobName string    `json:"kubernetes_job_name"`
	ConfigMapName     string    `json:"config_map_name"`
	Namespace         string    `json:"namespace"`
	Image             string    `json:"image"`
	Command           Command   `json:"command"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time,omitempty"`
	Status            JobStatus `json:"status"`
}

// JobStatus represents the status of a job execution
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCanceled  JobStatus = "canceled"
)

// StreamConfig represents stream configuration
type StreamConfig struct {
	StreamName          string     `json:"stream_name"`
	SyncMode            string     `json:"sync_mode"`
	CursorField         []string   `json:"cursor_field,omitempty"`
	PrimaryKey          [][]string `json:"primary_key,omitempty"`
	DestinationSyncMode string     `json:"destination_sync_mode"`
	Selected            bool       `json:"selected"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	RecordsSynced    int64                  `json:"records_synced"`
	BytesSynced      int64                  `json:"bytes_synced"`
	StreamsProcessed int                    `json:"streams_processed"`
	State            map[string]interface{} `json:"state"`
	Success          bool                   `json:"success"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	Duration         time.Duration          `json:"duration"`
}
