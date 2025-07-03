package telemetry

import "time"

// Telemetry constants
const (
	TelemetryUserIDFile      = "user_id"
	OlakeVersion             = "0.0.4"
	IPNotFound               = "NA"
	TelemetryConfigTimeout   = 30 * time.Second
	TelemetrySegmentAPIKey   = "AiWKKeaOKQvsOotHj5iGANpNhYG6OaM3"
	IPUrl                    = "https://api.ipify.org?format=text"
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
