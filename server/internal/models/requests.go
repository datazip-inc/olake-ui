package models

// LoginRequest represents the expected JSON structure for login requests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
