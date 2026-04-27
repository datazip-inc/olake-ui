package models

import (
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// BaseModel with common fields
type BaseModel struct {
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

// User represents the user entity
type User struct {
	BaseModel
	ID       int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Username string `json:"username" gorm:"column:username;size:100;uniqueIndex"`
	Password string `json:"password" gorm:"column:password;size:100"`
	Email    string `json:"email" gorm:"column:email;size:100;uniqueIndex"`
}

func (u *User) TableName() string {
	return constants.TableNameMap[constants.UserTable]
}

// ProjectSettings stores configuration scoped per project.
type ProjectSettings struct {
	BaseModel
	ID              int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID       string `json:"project_id" gorm:"column:project_id;size:255;uniqueIndex"`
	WebhookAlertURL string `json:"webhook_alert_url" gorm:"column:webhook_alert_url;size:512"`
}

func (s *ProjectSettings) TableName() string {
	return constants.TableNameMap[constants.ProjectSettingsTable]
}

// Source entity referencing User for auditing fields
type Source struct {
	BaseModel
	ID          int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"column:name;size:255"`
	ProjectID   string `json:"project_id" gorm:"column:project_id;size:255"`
	Config      string `json:"config" gorm:"column:config;type:jsonb"`
	Version     string `json:"version" gorm:"column:version;size:255"`
	Type        string `json:"type" gorm:"column:type;size:255"`
	CreatedByID int    `json:"-" gorm:"column:created_by_id"`
	UpdatedByID int    `json:"-" gorm:"column:updated_by_id"`

	CreatedBy *User `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;references:ID"`
	UpdatedBy *User `json:"updated_by,omitempty" gorm:"foreignKey:UpdatedByID;references:ID"`
}

func (s *Source) TableName() string {
	return constants.TableNameMap[constants.SourceTable]
}

// Destination entity referencing User
type Destination struct {
	BaseModel
	ID          int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"column:name;size:255"`
	ProjectID   string `json:"project_id" gorm:"column:project_id;size:255"`
	DestType    string `json:"type" gorm:"column:dest_type;size:255"`
	Version     string `json:"version" gorm:"column:version;size:255"`
	Config      string `json:"config" gorm:"column:config;type:jsonb"`
	CreatedByID int    `json:"-" gorm:"column:created_by_id"`
	UpdatedByID int    `json:"-" gorm:"column:updated_by_id"`

	CreatedBy *User `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;references:ID"`
	UpdatedBy *User `json:"updated_by,omitempty" gorm:"foreignKey:UpdatedByID;references:ID"`
}

func (d *Destination) TableName() string {
	return constants.TableNameMap[constants.DestinationTable]
}

// TODO_BEFORE_MERGE: confirm with team if we need to perform hard/soft delete
// Job represents a synchronization job
type Job struct {
	BaseModel
	ID               int     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name             string  `json:"name" gorm:"column:name;size:100"`
	SourceID         int     `json:"source_id" gorm:"column:source_id"`
	DestID           int     `json:"dest_id" gorm:"column:dest_id"`
	Active           bool    `json:"active" gorm:"column:active"`
	Frequency        string  `json:"frequency" gorm:"column:frequency;size:255"`
	StreamsConfig    string  `json:"streams_config" gorm:"column:streams_config;type:jsonb"`
	State            string  `json:"state" gorm:"column:state;type:jsonb"`
	AdvancedSettings *string `json:"advanced_settings" gorm:"column:advanced_settings;type:jsonb"`
	CreatedByID      int     `json:"-" gorm:"column:created_by_id"`
	UpdatedByID      int     `json:"-" gorm:"column:updated_by_id"`
	ProjectID        string  `json:"project_id" gorm:"column:project_id;size:255"`

	Source      *Source      `json:"source,omitempty" gorm:"foreignKey:SourceID;references:ID"`
	Destination *Destination `json:"destination,omitempty" gorm:"foreignKey:DestID;references:ID"`
	CreatedBy   *User        `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;references:ID"`
	UpdatedBy   *User        `json:"updated_by,omitempty" gorm:"foreignKey:UpdatedByID;references:ID"`
}

func (j *Job) TableName() string {
	return constants.TableNameMap[constants.JobTable]
}

type Catalog struct {
	BaseModel
	ID      int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Type    string `json:"type" gorm:"column:type;size:50"`
	Name    string `json:"name" gorm:"column:name;size:100"`
	Specs   string `json:"specs" gorm:"column:specs;type:jsonb"`
	Version string `json:"version" gorm:"column:version;size:255"`
}

func (c *Catalog) TableName() string {
	return constants.TableNameMap[constants.CatalogTable]
}

type Session struct {
	SessionKey    string    `json:"session_key" gorm:"column:session_key;primaryKey;size:64"`
	SessionData   []byte    `json:"session_data" gorm:"column:session_data;type:bytea"`
	SessionExpiry time.Time `json:"session_expiry" gorm:"column:session_expiry"`
}

func (s *Session) TableName() string {
	return constants.TableNameMap[constants.SessionTable]
}
