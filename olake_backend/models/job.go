package models

import "time"

type Job struct {
	JobID         int       `json:"job_id" orm:"column(job_id);pk;auto"`
	SourceID      int       `json:"source_id" orm:"column(source_id)"`
	DestID        int       `json:"dest_id" orm:"column(dest_id)"`
	Name          string    `json:"name" orm:"size(100)"`
	Connected     bool      `json:"connected"`
	CreatedAt     time.Time `json:"created_at" orm:"type(datetime)"`
	UpdatedAt     time.Time `json:"updated_at" orm:"type(datetime)"`
	LastRunTime   *time.Time `json:"last_run_time" orm:"type(datetime)"`
	LastRunState  string    `json:"last_run_state" orm:"size(100)"`
	Frequency     string    `json:"frequency" orm:"size(50)"`
	Config        string    `json:"config" orm:"size(500)"`
	CreatedBy     string    `json:"created_by"`
	UpdatedBy     string    `json:"updated_by"`
	DeletedAt     *time.Time `json:"deleted_at" orm:"null;type(datetime)"`
}

// GetLastSyncState returns the LastSyncState object
func (j *Job) GetLastSyncState() map[string]interface{} {
	return map[string]interface{}{
		"last_run_time":  j.LastRunTime,
		"last_run_state": j.LastRunState,
	}
}
