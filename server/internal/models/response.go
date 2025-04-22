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
type SourceTestConnectionResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Data    SourceTestConnectionRequest `json:"data"`
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
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Name string `json:"name"`
	} `json:"data"`
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
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Name string `json:"name"`
	} `json:"data"`
}
