package models

// LoginRequest represents the expected JSON structure for login requests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type SpecRequest struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}
type SourceTestConnectionRequest struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type DestinationTestConnectionRequest struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type CreateSourceRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type UpdateSourceRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type CreateDestinationRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
type UpdateDestinationRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Config  string `json:"config" orm:"type(jsonb)"`
}
