package types

import "time"

// TemporalConfig contains Temporal-related settings
type TemporalConfig struct {
	Address   string        `json:"address"`
	TaskQueue string        `json:"task_queue"`
	Timeout   time.Duration `json:"timeout"`
}