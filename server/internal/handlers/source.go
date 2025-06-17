package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/services"
	"github.com/datazip/olake-frontend/server/utils"
)

type SourceHandler struct {
	web.Controller
	sourceService *services.SourceService
}

func (c *SourceHandler) Prepare() {
	var err error
	c.sourceService, err = services.NewSourceService()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to initialize source service")
		return
	}
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	projectIDStr := c.Ctx.Input.Param(":projectid")

	sources, err := c.sourceService.GetAllSources(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	utils.SuccessResponse(&c.Controller, sources)
}

// @router /project/:projectid/sources [post]
func (c *SourceHandler) CreateSource() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req models.CreateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get user ID from session
	var userID *int
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if uid, ok := sessionUserID.(int); ok {
			userID = &uid
		}
	}

	if err := c.sourceService.CreateSource(req, projectID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create source: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	id, err := c.getIDFromPath()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	var req models.UpdateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get user ID from session
	var userID *int
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if uid, ok := sessionUserID.(int); ok {
			userID = &uid
		}
	}

	if err := c.sourceService.UpdateSource(id, req, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to update source: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id, err := c.getIDFromPath()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	response, err := c.sourceService.DeleteSource(id)
	if err != nil {
		if err.Error() == "source not found: " {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		} else {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to delete source: %s", err))
		}
		return
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	result, err := c.sourceService.TestConnection(req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to test connection: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, result)
}

// @router /sources/streams[post]
func (c *SourceHandler) GetSourceCatalog() {
	var req models.CreateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	catalog, err := c.sourceService.GetSourceCatalog(req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, catalog)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	id, err := c.getIDFromPath()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	jobs, err := c.sourceService.GetSourceJobs(id)
	if err != nil {
		if err.Error() == "source not found: " {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		} else {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get jobs by source ID")
		}
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"jobs": jobs,
	})
}

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	sourceType := c.GetString("type")

	versions, err := c.sourceService.GetSourceVersions(sourceType)
	if err != nil {
		if err.Error() == "source type is required" {
			utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Source type is required")
		} else {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get Docker versions")
		}
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version": versions,
	})
}

// Helper method to extract ID from path
func (c *SourceHandler) getIDFromPath() (int, error) {
	idStr := c.Ctx.Input.Param(":id")
	return strconv.Atoi(idStr)
}

// @router /project/:projectid/sources/spec [get]
func (c *SourceHandler) GetProjectSourceSpec() {
	_ = c.Ctx.Input.Param(":projectid")

	var req models.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	var spec interface{}

	switch req.Type {
	case "postgres":
		spec = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"host": map[string]interface{}{
					"type":        "string",
					"title":       "Postgres Host",
					"description": "Database host addresses for connection",
					"order":       1,
				},
				"port": map[string]interface{}{
					"type":        "integer",
					"title":       "Postgres Port",
					"description": "Database server listening port",
					"order":       2,
				},
				"database": map[string]interface{}{
					"type":        "string",
					"title":       "Database Name",
					"description": "The name of the database to use for the connection",
					"order":       3,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "Username used to authenticate with the database",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "Password for database authentication",
					"format":      "password",
					"order":       5,
				},
				"jdbc_url_params": map[string]interface{}{
					"type":        "string",
					"title":       "JDBC URL Parameters",
					"description": "Additional JDBC URL parameters for connection tuning (optional)",
					"order":       6,
				},
				"ssl": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"mode": map[string]interface{}{
							"type":        "string",
							"title":       "SSL Mode",
							"description": "Database connection SSL configuration (e.g., SSL mode)",
							"enum":        []string{"disable", "require", "verify-ca", "verify-full"},
						},
					},
					"order": 7,
				},
				"update_method": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"replication_slot": map[string]interface{}{
							"type":        "string",
							"title":       "Replication Slot",
							"description": "Slot to retain WAL logs for consistent replication",
						},
						"intial_wait_time": map[string]interface{}{
							"type":        "integer",
							"title":       "Initial Wait Time",
							"description": "Idle timeout for WAL log reading",
						},
					},
					"order": 8,
				},
				"reader_batch_size": map[string]interface{}{
					"type":        "integer",
					"title":       "Reader Batch Size",
					"description": "Maximum batch size for read operations",
					"order":       9,
				},
				"default_mode": map[string]interface{}{
					"type":        "string",
					"title":       "Default Mode",
					"description": "Default sync mode (e.g., CDC — Change Data Capture OR Full_Refresh)",
					"order":       10,
				},
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Max parallel threads for chunk snapshotting",
					"order":       11,
				},
			},
			"required": []string{"host", "port", "database", "username", "password"},
		}

	case "mysql":
		spec = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"hosts": map[string]interface{}{
					"type":        "string",
					"title":       "MySQL Host",
					"description": "Database host addresses for connection",
					"order":       1,
				},
				"port": map[string]interface{}{
					"type":        "integer",
					"title":       "Port",
					"description": "Database server listening port",
					"order":       2,
				},
				"database": map[string]interface{}{
					"type":        "string",
					"title":       "Database",
					"description": "The name of the database to use for the connection",
					"order":       3,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "Username used to authenticate with the database",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "Password for database authentication",
					"format":      "password",
					"order":       5,
				},
				"update_method": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"intial_wait_time": map[string]interface{}{
							"type":        "integer",
							"title":       "Initial Wait Time",
							"description": "Idle timeout for Bin log reading",
						},
					},
					"order": 6,
				},
				"default_mode": map[string]interface{}{
					"type":        "string",
					"title":       "Default Mode",
					"description": "Default sync mode (e.g., CDC — Change Data Capture OR Full_Refresh)",
					"order":       7,
				},
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Maximum concurrent threads for data sync",
					"order":       8,
				},
				"backoff_retry_count": map[string]interface{}{
					"type":        "integer",
					"title":       "Backoff Retry Count",
					"description": "Number of sync retries (exponential backoff on failure)",
					"order":       9,
				},
				"tls_skip_verify": map[string]interface{}{
					"type":        "boolean",
					"title":       "Skip TLS Verification",
					"description": "Determines if TLS certificate verification should be skipped for secure connections",
					"order":       10,
				},
			},
			"required": []string{"hosts", "username", "password", "database", "port"},
		}

	case "mongodb":
		spec = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"hosts": map[string]interface{}{
					"type":        "array",
					"title":       "Hosts",
					"description": "Specifies the hostnames or IP addresses of MongoDB for connection",
					"items":       map[string]interface{}{"type": "string"},
					"order":       1,
				},
				"database": map[string]interface{}{
					"type":        "string",
					"title":       "Database Name",
					"description": "The name of the MongoDB database selected for replication",
					"order":       2,
				},
				"authdb": map[string]interface{}{
					"type":        "string",
					"title":       "Auth DB",
					"description": "Authentication database (mostly:admin)",
					"order":       3,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "Username for MongoDB authentication",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "Password with the username for authentication",
					"format":      "password",
					"order":       5,
				},
				"replica-set": map[string]interface{}{
					"type":        "string",
					"title":       "Replica Set",
					"description": "MongoDB replica set name (if applicable)",
					"order":       6,
				},
				"read-preference": map[string]interface{}{
					"type":        "string",
					"title":       "Read Preference",
					"description": "Read preference for MongoDB (e.g., secondaryPreferred)",
					"order":       7,
				},
				"srv": map[string]interface{}{
					"type":        "boolean",
					"title":       "Use SRV",
					"description": "Enable this option if using DNS SRV connection strings. When set to true, the hosts field must contain only one entry — a DNS SRV address ([\"mongodatatest.pigiy.mongodb.net\"])",
					"order":       8,
				},
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Max parallel threads for chunk snapshotting",
					"order":       9,
				},
				"default_mode": map[string]interface{}{
					"type":        "string",
					"title":       "Default Mode",
					"description": "Default sync mode (e.g., CDC — Change Data Capture OR Full_Refresh)",
					"order":       10,
				},
				"backoff_retry_count": map[string]interface{}{
					"type":        "integer",
					"title":       "Retry Count",
					"description": "Number of sync retry attempts using exponential backoff",
					"order":       11,
				},
				"partition_strategy": map[string]interface{}{
					"type":        "string",
					"title":       "Chunking Strategy",
					"description": "Chunking Strategy (timestamp, uses splitVector strategy if the field is left empty)",
					"order":       12,
				},
			},
			"required": []string{"hosts", "username", "password", "database", "authdb"},
		}

	default:
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Unsupported source type")
		return
	}

	utils.SuccessResponse(&c.Controller, models.SpecResponse{
		Version: req.Version,
		Type:    req.Type,
		Spec:    spec,
	})
}
