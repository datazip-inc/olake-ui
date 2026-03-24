package models

type CatalogRequest struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	OptimizerGroup  string            `json:"optimizerGroup,omitempty"`
	TableFormatList []string          `json:"tableFormatList"`
	StorageConfig   map[string]string `json:"storageConfig"`
	AuthConfig      map[string]string `json:"authConfig"`
	Properties      map[string]string `json:"properties"`
	TableProperties map[string]string `json:"tableProperties"`
}

// LogInfo represents the log information from terminal execution
type LogInfo struct {
	LogStatus string   `json:"logStatus"`
	Logs      []string `json:"logs"`
}

type SetTablePropertiesRequest struct {
	Catalog    string            `json:"catalog"`
	Database   string            `json:"database"`
	Table      string            `json:"table"`
	Properties map[string]string `json:"properties"`
}

// SetTablePropertiesResponse represents the response from setting table properties
type SetTablePropertiesResponse struct {
	SessionID string   `json:"sessionId"`
	Status    string   `json:"status"`
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Logs      []string `json:"logs,omitempty"`
}

// TerminalExecuteRequest represents the request body for terminal SQL execution
type TerminalExecuteRequest struct {
	SQL string `json:"sql"`
}

// TerminalSessionResponse represents the response from terminal execute
type TerminalSessionResponse struct {
	SessionID string `json:"sessionId"`
}

type SQLInput struct {
	MinorTriggerInterval   string `json:"minorTriggerInterval"`
	MajorTriggerInterval   string `json:"majorTriggerInterval"`
	FullTriggerInterval    string `json:"fullTriggerInterval"`
	TargetFileSize         int64  `json:"targetFileSize"`
	EnabledForOptimisation string `json:"enabledForOptimisation"`
}
