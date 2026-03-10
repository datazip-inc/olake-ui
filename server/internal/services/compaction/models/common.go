package models

// CatalogRequest represents the request to create or update a catalog
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

// CatalogResponse represents the response from catalog operations
type CatalogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CatalogsResponse represents catalogs with their databases (no table details)
type CatalogsResponse struct {
	Catalogs []CatalogWithDatabases `json:"catalogs"`
}

type CatalogWithDatabases struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Databases []string `json:"databases"`
}

// TablesResponse represents tables with full details for a specific catalog/database
type TablesResponse struct {
	Catalog  string      `json:"catalog"`
	Database string      `json:"database"`
	Tables   []TableInfo `json:"tables"`
}

type TableInfo struct {
	Name      string            `json:"name"`
	TotalSize string            `json:"totalSize,omitempty"`
	ByOLake   bool              `json:"byOLake"`
	Major     *OptimizationInfo `json:"major"`
	Minor     *OptimizationInfo `json:"minor"`
	Full      *OptimizationInfo `json:"full"`
	Enabled   bool              `json:"enabled"`
}

type OptimizationInfo struct {
	LastRun string `json:"last-run,omitempty"`
	Status  string `json:"status,omitempty"`
}

// SetTablePropertiesRequest represents the request to set table properties
type SetTablePropertiesRequest struct {
	Catalog    string            `json:"catalog"`
	Database   string            `json:"database"`
	Table      string            `json:"table"`
	Properties map[string]string `json:"properties"`
}

// CompactionCronConfigRequest represents the request to set compaction cron configuration
// TriggerInterval values are in milliseconds. Use "never" to disable a specific compaction type.
type CompactionCronConfigRequest struct {
	Enabled              bool   `json:"enabled"`
	MinorTriggerInterval string `json:"minorTriggerInterval"`
	MajorTriggerInterval string `json:"majorTriggerInterval"`
	FullTriggerInterval  string `json:"fullTriggerInterval"`
}

// CompactionCronConfigResponse represents the response from setting cron configuration
type CompactionCronConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
