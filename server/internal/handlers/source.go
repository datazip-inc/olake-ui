package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	c.sourceService, err = services.NewSourceService()
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to initialize source service", err)
		return
	}
	c.Ctx.Request = c.Ctx.Request.WithContext(ctx)
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	projectID := c.Ctx.Input.Param(":projectid")
	sources, err := c.sourceService.GetAllSources(context.Background(), projectID)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources", err)
		return
	}
	utils.SuccessResponse(&c.Controller, sources)
}

// @router /project/:projectid/sources [post]
func (c *SourceHandler) CreateSource() {
	projectID := c.Ctx.Input.Param(":projectid")
	var req models.CreateSourceRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.sourceService.CreateSource(context.Background(), req, projectID, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create source", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	id := GetIDFromPath(&c.Controller)
	var req models.UpdateSourceRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	userID := GetUserIDFromSession(&c.Controller)
	if err := c.sourceService.UpdateSource(context.Background(), id, req, userID); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update source", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id := GetIDFromPath(&c.Controller)
	resp, err := c.sourceService.DeleteSource(context.Background(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			respondWithError(&c.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete source", err)
		}
		return
	}
	utils.SuccessResponse(&c.Controller, resp)
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	result, err := c.sourceService.TestConnection(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to test connection", err)
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /sources/streams[post]
func (c *SourceHandler) GetSourceCatalog() {
	var req models.StreamsRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	catalog, err := c.sourceService.GetSourceCatalog(context.Background(), req)
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source catalog", err)
		return
	}
	utils.SuccessResponse(&c.Controller, catalog)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	id := GetIDFromPath(&c.Controller)
	jobs, err := c.sourceService.GetSourceJobs(context.Background(), id)
	if err != nil {
		if errors.Is(err, constants.ErrSourceNotFound) {
			respondWithError(&c.Controller, http.StatusNotFound, "Source not found", err)
		} else {
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get source jobs", err)
		}
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"jobs": jobs})
}

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	sourceType := c.GetString("type")
	versions, err := c.sourceService.GetSourceVersions(context.Background(), sourceType)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to get Docker versions"
		if err.Error() == "source type is required" {
			status = http.StatusBadRequest
			msg = "Source type is required"
		}
		respondWithError(&c.Controller, status, msg, err)
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{"version": versions})
}

// @router /project/:projectid/sources/spec [get]
func (c *SourceHandler) GetProjectSourceSpec() {
	_ = c.Ctx.Input.Param(":projectid")

	var req models.SpecRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
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
					"type":        "obj",
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
	case "oracle":
		spec = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"host": map[string]interface{}{
					"type":        "string",
					"title":       "Host",
					"description": "Hostname or IP address of the Oracle database server.",
					"order":       1,
				},
				"port": map[string]interface{}{
					"type":        "integer",
					"title":       "Port",
					"description": "Port number on which the Oracle database is listening.",
					"order":       2,
				},
				"service_name": map[string]interface{}{
					"type":        "string",
					"title":       "Service Name",
					"description": "The Oracle database service name to connect to.",
					"order":       3,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "Username for authenticating with the Oracle database.",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "Password for the Oracle database user.",
					"format":      "password",
					"order":       5,
				},
				"jdbc_url_params": map[string]interface{}{
					"type":        "obj",
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
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Max parallel threads for chunk snapshotting",
					"order":       8,
				},
				"backoff_retry_count": map[string]interface{}{
					"type":        "integer",
					"title":       "Retry Count",
					"description": "Number of sync retry attempts using exponential backoff",
					"order":       9,
				},
				"sid": map[string]interface{}{
					"type":        "string",
					"title":       "SID",
					"description": "The Oracle database SID to connect to.",
					"order":       10,
				},
			},
			"required": []string{"host", "port", "service_name", "username", "password"},
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
