package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/internal/temporal"
	"github.com/datazip/olake-frontend/server/utils"
)

type SourceHandler struct {
	web.Controller
	sourceORM  *database.SourceORM
	userORM    *database.UserORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func (c *SourceHandler) Prepare() {
	c.sourceORM = database.NewSourceORM()
	c.userORM = database.NewUserORM()
	c.jobORM = database.NewJobORM()

	// Initialize Temporal client
	var err error
	c.tempClient, err = temporal.NewClient()
	if err != nil {
		logs.Error("Failed to create Temporal client: %v", err)
	}
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	projectIDStr := c.Ctx.Input.Param(":projectid")
	sourceItems := make([]models.SourceDataItem, 0, len(sources))

	for _, source := range sources {
		item := models.SourceDataItem{
			ID:        source.ID,
			Name:      source.Name,
			Type:      source.Type,
			Version:   source.Version,
			Config:    source.Config,
			CreatedAt: source.CreatedAt.Format(time.RFC3339),
			UpdatedAt: source.UpdatedAt.Format(time.RFC3339),
		}

		setUsernames(&item.CreatedBy, &item.UpdatedBy, source.CreatedBy, source.UpdatedBy)

		jobs, err := c.jobORM.GetBySourceID(source.ID)
		var success bool
		item.Jobs, success = buildJobDataItems(jobs, err, projectIDStr, "source", c.tempClient, &c.Controller)
		if !success {
			return // Error occurred in buildJobDataItems
		}

		sourceItems = append(sourceItems, item)
	}

	utils.SuccessResponse(&c.Controller, sourceItems)
}

// @router /project/:projectid/sources [post]
func (c *SourceHandler) CreateSource() {
	var req models.CreateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert request to Source model
	source := &models.Source{
		Name:    req.Name,
		Type:    req.Type,
		Version: req.Version,
		Config:  req.Config,
	}

	// Get project ID if needed
	source.ProjectID = c.Ctx.Input.Param(":projectid")

	// Set created by if user is logged in
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user, err := c.userORM.GetByID(userID.(int))
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get user")
			return
		}
		source.CreatedBy = user
		source.UpdatedBy = user
	}
	if err := c.sourceORM.Create(source); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create source: %s", err))
		return
	}

	// Track source creation event
	go func() {
		if err := telemetry.TrackSourceCreation(
			c.Ctx.Request.Context(),
			source.ID,
			source.Name,
			source.Type,
			source.Version,
			source.CreatedAt,
		); err != nil {
			logs.Error("Failed to track source creation event: %v", err)
		}
	}()

	// Track sources status after creation
	go func() {
		if err := telemetry.TrackSourcesStatus(c.Ctx.Request.Context(), userID); err != nil {
			logs.Error("Failed to track sources status: %v", err)
		}
	}()

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	id := GetIDFromPath(&c.Controller)
	var req models.UpdateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	// Get existing source
	existingSource, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Update fields
	existingSource.Name = req.Name
	existingSource.Config = req.Config
	existingSource.Type = req.Type
	existingSource.Version = req.Version
	existingSource.UpdatedAt = time.Now()

	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		existingSource.UpdatedBy = user
	}

	if err := c.sourceORM.Update(existingSource); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update source")
		return
	}

	// Track sources status after update
	go func() {
		if err := telemetry.TrackSourcesStatus(c.Ctx.Request.Context(), userID); err != nil {
			logs.Error("Failed to track sources status: %v", err)
		}
	}()

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	id := GetIDFromPath(&c.Controller)
	source, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Get all jobs using this source
	jobs, err := c.jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get jobs for source")
		return
	}

	// Get user ID before deletion for telemetry
	userID := c.GetSession(constants.SessionUserID)

	// Deactivate all jobs using this source
	for _, job := range jobs {
		job.Active = false
		if err := c.jobORM.Update(job); err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to deactivate jobs using this source")
			return
		}
	}

	// Delete the source
	if err := c.sourceORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}

	// Track sources status after deletion
	go func() {
		if err := telemetry.TrackSourcesStatus(c.Ctx.Request.Context(), userID); err != nil {
			logs.Error("Failed to track sources status: %v", err)
		}
	}()

	utils.SuccessResponse(&c.Controller, &models.DeleteSourceResponse{
		Name: source.Name,
	})
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	result, err := c.tempClient.TestConnection(context.Background(), "config", req.Type, req.Version, req.Config)
	if result == nil {
		result = map[string]interface{}{
			"message": err.Error(),
			"status":  "failed",
		}
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /sources/streams[post]
func (c *SourceHandler) GetSourceCatalog() {
	var req models.StreamsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	oldStreams := ""
	// Load job details if JobID is provided
	if req.JobID >= 0 {
		job, err := c.jobORM.GetByID(req.JobID)
		if err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Job not found")
			return
		}
		oldStreams = job.StreamsConfig
	}

	// Use Temporal client to get the catalog
	var newStreams map[string]interface{}
	var err error
	if c.tempClient != nil {
		newStreams, err = c.tempClient.GetCatalog(
			c.Ctx.Request.Context(),
			req.Type,
			req.Version,
			req.Config,
			oldStreams,
		)
	}
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to get catalog: %v", err))
		return
	}
	utils.SuccessResponse(&c.Controller, newStreams)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	id := GetIDFromPath(&c.Controller)
	// Check if source exists
	_, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Create a job ORM and get jobs by source ID
	jobs, err := c.jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get jobs by source ID")
		return
	}
	// Format as required by API contract
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"jobs": jobs,
	})
}

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	// Get source type from query parameter
	sourceType := c.GetString("type")
	if sourceType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Source type is required")
		return
	}

	// Get versions from Docker Hub
	imageName := fmt.Sprintf("olakego/source-%s", sourceType)

	versions, err := utils.GetDockerHubTags(imageName)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get Docker versions")
		return
	}
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version": versions,
	})
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
