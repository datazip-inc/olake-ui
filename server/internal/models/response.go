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
type SpecResponse struct {
	Version string `json:"version"`
	Type    string `json:"type"`
	Spec    string `json:"spec" orm:"type(jsonb)"`
}
type DestinationTestConnectionResponse struct {
	Success bool                             `json:"success"`
	Message string                           `json:"message"`
	Data    DestinationTestConnectionRequest `json:"data"`
}
type CreateSourceResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    CreateSourceRequest `json:"data"`
}
type UpdateSourceResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    UpdateSourceRequest `json:"data"`
}
type DeleteSourceResponse struct {
	Name string `json:"name"`
}
type CreateDestinationResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    CreateDestinationRequest `json:"data"`
}
type UpdateDestinationResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    UpdateDestinationRequest `json:"data"`
}
type DeleteDestinationResponse struct {
	Name string `json:"name"`
}

// JobResponse represents a job in response format
type JobResponse struct {
	ID            int                  `json:"id"`
	Name          string               `json:"name"`
	Source        JobSourceConfig      `json:"source"`
	Destination   JobDestinationConfig `json:"destination"`
	StreamsConfig string               `json:"streams_config"`
	Frequency     string               `json:"frequency"`
	LastRunTime   string               `json:"last_run_time,omitempty"`
	LastRunState  string               `json:"last_run_state,omitempty"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
	Activate      bool                 `json:"activate"`
	CreatedBy     string               `json:"created_by,omitempty"`
	UpdatedBy     string               `json:"updated_by,omitempty"`
}

type CreateJobResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    CreateJobRequest `json:"data"`
}

type UpdateJobResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    UpdateJobRequest `json:"data"`
}

type DeleteJobResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Name string `json:"name"`
	} `json:"data"`
}

type GetJobStreamsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		StreamsConfig string `json:"streams_config"`
	} `json:"data"`
}

// SourceDataItem represents a single source in the response data list
type SourceDataItem struct {
	ID        int                      `json:"id"`
	Name      string                   `json:"name"`
	Type      string                   `json:"type"`
	Version   string                   `json:"version"`
	Config    string                   `json:"config"`
	CreatedAt string                   `json:"created_at"`
	UpdatedAt string                   `json:"updated_at"`
	CreatedBy string                   `json:"created_by"` // only username of user
	UpdatedBy string                   `json:"updated_by"` // only username of user
	Jobs      []map[string]interface{} `json:"jobs"`
}

// DestinationDataItem represents a single destination in the response data list
type DestinationDataItem struct {
	ID        int                      `json:"id"`
	Name      string                   `json:"name"`
	Type      string                   `json:"type"`
	Version   string                   `json:"version"`
	Config    string                   `json:"config"`
	CreatedAt string                   `json:"created_at"`
	UpdatedAt string                   `json:"updated_at"`
	CreatedBy string                   `json:"created_by"` // only username of user
	UpdatedBy string                   `json:"updated_by"` // only username of user
	Jobs      []map[string]interface{} `json:"jobs"`
}
