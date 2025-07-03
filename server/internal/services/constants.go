package services

import "errors"

// Common error messages
var (
	// User related errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrPasswordProcessing = errors.New("failed to process password")

	// Source related errors
	ErrSourceNotFound     = errors.New("source not found")
	ErrSourceTypeRequired = errors.New("source type is required")
	ErrSourceUpdateFailed = errors.New("failed to update source")

	// Destination related errors
	ErrDestinationNotFound        = errors.New("destination not found")
	ErrDestinationTypeRequired    = errors.New("destination type is required")
	ErrDestinationVersionRequired = errors.New("destination version is required")
	ErrDestinationUpdateFailed    = errors.New("failed to update destination")

	// Job related errors
	ErrJobNotFound     = errors.New("job not found")
	ErrJobCreateFailed = errors.New("failed to create job")
	ErrJobUpdateFailed = errors.New("failed to update job")

	// Temporal related errors
	ErrTemporalClientNotAvailable = errors.New("temporal client not available")
	ErrWorkflowExecutionFailed    = errors.New("temporal workflow execution failed")

	// General errors
	ErrFailedToRetrieve = errors.New("failed to retrieve")
	ErrFailedToProcess  = errors.New("failed to process")
	ErrFailedToCreate   = errors.New("failed to create")
	ErrFailedToUpdate   = errors.New("failed to update")
	ErrFailedToDelete   = errors.New("failed to delete")
)

// Error message formats
const (
	// Format strings for error messages with dynamic values
	ErrFormatFailedToFindUser          = "failed to find user: %s"
	ErrFormatFailedToGetUser           = "failed to get user: %s"
	ErrFormatFailedToGetCatalog        = "failed to get catalog: %s"
	ErrFormatFailedToGetJobs           = "failed to get jobs: %s"
	ErrFormatFailedToDeactivateJob     = "failed to deactivate job %d: %s"
	ErrFormatFailedToGetDockerVersions = "failed to get Docker versions: %s"
	ErrFormatFailedToRetrieveJobs      = "failed to retrieve jobs: %s"
)

// Success messages
const (
	MsgUserCreated   = "User created successfully"
	MsgUserUpdated   = "User updated successfully"
	MsgUserDeleted   = "User deleted successfully"
	MsgSourceCreated = "Source created successfully"
	MsgSourceUpdated = "Source updated successfully"
	MsgSourceDeleted = "Source deleted successfully"
	MsgDestCreated   = "Destination created successfully"
	MsgDestUpdated   = "Destination updated successfully"
	MsgDestDeleted   = "Destination deleted successfully"
	MsgJobCreated    = "Job created successfully"
	MsgJobUpdated    = "Job updated successfully"
	MsgJobDeleted    = "Job deleted successfully"
)

// Validation messages
const (
	ValidationEmailRequired    = "Email is required"
	ValidationPasswordRequired = "Password is required"
	ValidationNameRequired     = "Name is required"
)
