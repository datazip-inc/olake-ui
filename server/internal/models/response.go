package models

// LoginResponse represents the JSON structure sent back to the client
type LoginResponse struct {
	Message string `json:"message"` // Human-readable message
	Success bool   `json:"success"` // Indicates if login was successful
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
