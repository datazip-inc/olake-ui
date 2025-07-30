package models

import (
	"time"

	"github.com/datazip/olake-ui/server/internal/constants"
)

// BaseModel with common fields
type BaseModel struct {
	CreatedAt time.Time  `json:"created_at" orm:"column(created_at);auto_now_add;type(datetime)"`
	UpdatedAt time.Time  `json:"updated_at" orm:"column(updated_at);auto_now;type(datetime)"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" orm:"column(deleted_at);null;type(datetime)"`
}

// User represents the user entity
type User struct {
	BaseModel `orm:"embedded"`
	ID        int    `json:"id" orm:"column(id);pk;auto"`
	Username  string `json:"username" orm:"size(100);unique"`
	Password  string `json:"password" orm:"size(100)"` // Hidden in JSON
	Email     string `json:"email" orm:"size(100);unique"`
}

func (u *User) TableName() string {
	return constants.TableNameMap[constants.UserTable]
}

// Source entity referencing User for auditing fields
type Source struct {
	BaseModel `orm:"embedded"`
	ID        int    `json:"id" orm:"column(id);pk;auto"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id" orm:"column(project_id)"`
	Config    string `json:"config" orm:"type(jsonb)"`
	Version   string `json:"version"`
	CreatedBy *User  `json:"created_by" orm:"rel(fk)"`
	UpdatedBy *User  `json:"updated_by" orm:"rel(fk)"`
	Type      string `json:"type"`
}

func (s *Source) TableName() string {
	return constants.TableNameMap[constants.SourceTable]
}

// Destination entity referencing User
type Destination struct {
	BaseModel `orm:"embedded"`
	ID        int    `json:"id" orm:"column(id);pk;auto"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id" orm:"column(project_id)"`
	DestType  string `json:"type"`
	Version   string `json:"version"`
	Config    string `json:"config" orm:"type(jsonb)"`
	CreatedBy *User  `json:"created_by" orm:"rel(fk)"`
	UpdatedBy *User  `json:"updated_by" orm:"rel(fk)"`
}

func (d *Destination) TableName() string {
	return constants.TableNameMap[constants.DestinationTable]
}

// Job represents a synchronization job
type Job struct {
	BaseModel     `orm:"embedded"`
	ID            int          `json:"id" orm:"column(id);pk;auto"`
	Name          string       `json:"name" orm:"size(100)"`
	SourceID      *Source      `json:"source_id" orm:"column(source_id);rel(fk)"`
	DestID        *Destination `json:"dest_id" orm:"column(dest_id);rel(fk)"`
	Active        bool         `json:"active"`
	Frequency     string       `json:"frequency"`
	StreamsConfig string       `json:"streams_config" orm:"type(jsonb)"`
	State         string       `json:"state" orm:"type(jsonb)"`
	CreatedBy     *User        `json:"created_by" orm:"rel(fk)"`
	UpdatedBy     *User        `json:"updated_by" orm:"rel(fk)"`
	ProjectID     string       `json:"project_id" orm:"column(project_id)"`
}

func (j *Job) TableName() string {
	return constants.TableNameMap[constants.JobTable]
}

type Catalog struct {
	BaseModel `orm:"embedded"`
	ID        int    `json:"id" orm:"column(id);pk;auto"`
	Type      string `json:"type" orm:"size(50)"`
	Name      string `json:"name" orm:"size(100)"`
	Specs     string `json:"specs" orm:"type(jsonb)"`
	Version   string `json:"version"`
}

func (c *Catalog) TableName() string {
	return constants.TableNameMap[constants.CatalogTable]
}

type Session struct {
	SessionKey    string    `json:"session_key" orm:"column(session_key);pk;size(64)"`
	SessionData   string    `json:"session_data" orm:"column(session_data);type(text)"`
	SessionExpiry time.Time `json:"session_expiry" orm:"column(session_expiry)"`
}

func (s *Session) TableName() string {
	return constants.TableNameMap[constants.SessionTable]
}
