package models

import "time"

// User represents the user entity
type User struct {
	UserID    int      `json:"user_id" orm:"pk;auto"`
	Username  string    `json:"username" orm:"size(100);unique"`
	Password  string    `json:"password" orm:"size(100)"`
	Email     string    `json:"email" orm:"size(100);unique"`
	CreatedAt time.Time `json:"created_at" orm:"type(datetime)"`
	UpdatedAt time.Time `json:"updated_at" orm:"type(datetime)"`
	DeletedAt *time.Time `json:"deleted_at" orm:"null;type(datetime)"`
}