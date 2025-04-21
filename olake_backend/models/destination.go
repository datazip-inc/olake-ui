package models

import "time"

// Destination represents the destination entity
type Destination struct {
	DestID      int      `json:"dest_id" orm:"column(dest_id);pk;auto"`
	Name        string    `json:"name"`
	ProjectID   uint      `json:"project_id" orm:"column(project_id)"`
	Config      string    `json:"config"`
	CreatedAt   time.Time `json:"created_at" orm:"type(datetime)"`
	UpdatedAt   time.Time `json:"updated_at" orm:"type(datetime)"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
	DestType    string    `json:"dest_type"`
	DeletedAt   *time.Time `json:"deleted_at" orm:"null;type(datetime)"`
}