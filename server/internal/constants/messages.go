package constants

import "errors"

// Common error messages
var (
	// User related errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrPasswordProcessing = errors.New("failed to process password")

	// Source related errors
	ErrSourceNotFound      = errors.New("source not found")
	ErrDestinationNotFound = errors.New("destination not found")
	ErrJobNotFound         = errors.New("job not found")
)

// Validation messages
const (
	ValidationInvalidRequestFormat = "Invalid request format"
)
