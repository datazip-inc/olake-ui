package models

import "time"

type Source struct {
	SourceID    int      `json:"source_id" orm:"column(source_id);pk;auto"`
	Name        string    `json:"name"`
	ProjectID   uint      `json:"project_id" orm:"column(project_id)"`
	Config      string    `json:"config"`
	CreatedAt   time.Time `json:"created_at" orm:"type(datetime)"`
	UpdatedAt   time.Time `json:"updated_at" orm:"type(datetime)"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
	SourceType  string    `json:"source_type"`
	DeletedAt   *time.Time `json:"deleted_at" orm:"null;type(datetime)"`
}