package dto

type LoginResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SpecResponse struct {
	Version string      `json:"version"`
	Type    string      `json:"type"`
	Spec    interface{} `json:"spec" orm:"type(jsonb)"`
}
type SpecOutput struct {
	Spec map[string]interface{} `json:"spec"`
}

// Reuse generic API response with generics
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type DeleteSourceResponse struct {
	Name string `json:"name"`
}

type DeleteDestinationResponse struct {
	Name string `json:"name"`
}

type JobStatus struct {
	Activate bool `json:"activate"`
}

// JobNameCheckResponse
type CheckUniqueJobNameResponse struct {
	Unique bool `json:"unique"`
}

// TestConnectionResponse
type TestConnectionResponse struct {
	ConnectionResult map[string]interface{}   `json:"connection_result"`
	Logs             []map[string]interface{} `json:"logs"`
}

// Job response
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

type JobTask struct {
	Runtime   string `json:"runtime"`
	StartTime string `json:"start_time"`
	Status    string `json:"status"`
	FilePath  string `json:"file_path"`
}

type SourceDataItem struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	Version   string        `json:"version"`
	Config    string        `json:"config"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
	CreatedBy string        `json:"created_by"`
	UpdatedBy string        `json:"updated_by"`
	Jobs      []JobDataItem `json:"jobs"`
}

type DestinationDataItem struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	Version   string        `json:"version"`
	Config    string        `json:"config"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
	CreatedBy string        `json:"created_by"`
	UpdatedBy string        `json:"updated_by"`
	Jobs      []JobDataItem `json:"jobs"`
}

type JobDataItem struct {
	Name            string `json:"name"`
	ID              int    `json:"id"`
	Activate        bool   `json:"activate"`
	SourceName      string `json:"source_name,omitempty"`
	SourceType      string `json:"source_type,omitempty"`
	DestinationName string `json:"destination_name,omitempty"`
	DestinationType string `json:"destination_type,omitempty"`
	LastRunTime     string `json:"last_run_time,omitempty"`
	LastRunState    string `json:"last_run_state,omitempty"`
}
