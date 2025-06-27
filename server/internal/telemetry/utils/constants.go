package utils

import "time"

// Telemetry event constants
const (
	EventUserLogin           = "user_login"
	EventJobCreated          = "job_created"
	EventSyncStarted         = "sync_started"
	EventSyncCompleted       = "sync_completed"
	EventSyncFailed          = "sync_failed"
	EventSourceCreated       = "source_created"
	EventDestinationCreated  = "destination_created"
	EventSourcesUpdated      = "sources_updated"
	EventDestinationsUpdated = "destinations_updated"
)

// Telemetry configuration constants
const (
	TelemetryAnonymousIDFile       = "telemetry_id"
	TelemetryVersion               = "0.0.1"
	TelemetryIPNotFoundPlaceholder = "NA"
	TelemetryConfigTimeout         = time.Second
	TelemetrySegmentAPIKey         = "AiWKKeaOKQvsOotHj5iGANpNhYG6OaM3"
)
