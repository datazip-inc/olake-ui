package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
	"go.temporal.io/api/workflowservice/v1"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/internal/temporal"
	"github.com/datazip/olake-server/utils"
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
	tempAddress := web.AppConfig.DefaultString("TEMPORAL_ADDRESS", "localhost:7233")
	tempClient, err := temporal.NewClient(tempAddress)
	if err != nil {
		// Log the error but continue - we'll fall back to direct Docker execution if Temporal fails
		fmt.Printf("Failed to create Temporal client: %v\n", err)
	}
	c.tempClient = tempClient
	c.tempClient = tempClient
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Format response data
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

		// Add creator username if available
		if source.CreatedBy != nil {
			item.CreatedBy = source.CreatedBy.Username
		}

		// Add updater username if available
		if source.UpdatedBy != nil {
			item.UpdatedBy = source.UpdatedBy.Username
		}

		// Fetch associated jobs for this source
		jobs, err := c.jobORM.GetBySourceID(source.ID)
		sourceJobs := make([]models.JobDataItem, 0) // always initialize
		if err == nil {
			for _, job := range jobs {
				jobInfo := models.JobDataItem{
					Name:     job.Name,
					ID:       job.ID,
					Activate: job.Active,
				}
				// Add destination name if available
				if job.DestID != nil {
					jobInfo.DestinationName = job.DestID.Name
					jobInfo.DestinationType = job.DestID.DestType
				}

				query := fmt.Sprintf("WorkflowId between 'sync-%d-%d' and 'sync-%d-%d-~'", projectID, job.ID, projectID, job.ID)
				fmt.Println("Query:", query)
				// List workflows using the direct query
				resp, err := c.tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
					Query:    query,
					PageSize: 1,
				})
				if err != nil {
					utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to list workflows: %v", err))
					return
				}

				if len(resp.Executions) > 0 {
					jobInfo.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
					jobInfo.LastRunState = resp.Executions[0].Status.String()
				} else {
					jobInfo.LastRunTime = ""
					jobInfo.LastRunState = ""
				}

				sourceJobs = append(sourceJobs, jobInfo)
			}
		}
		// Assign jobs even if empty
		item.Jobs = sourceJobs

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
	projectIDStr := c.Ctx.Input.Param(":projectid")
	source.ProjectID = projectIDStr

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
	fmt.Printf("source: %+v\n", source)

	if err := c.sourceORM.Create(source); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create source: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [put]
func (c *SourceHandler) UpdateSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

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

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/sources/:id [delete]
func (c *SourceHandler) DeleteSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}
	name, err := c.sourceORM.GetName(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}
	jobs, err := c.jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get jobs by source ID")

	}
	for _, job := range jobs {
		job.Active = false
	}

	if err := c.sourceORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete source")
		return
	}

	utils.SuccessResponse(&c.Controller, &models.DeleteSourceResponse{
		Name: name,
	})
}

// @router /project/:projectid/sources/test [post]
func (c *SourceHandler) TestConnection() {
	var req models.SourceTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	result, _ := c.tempClient.TestConnection(context.Background(), req.Type, req.Version, req.Config)
	// if err != nil {
	// 	//utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to test connection")
	// 	return
	// }
	utils.SuccessResponse(&c.Controller, result)

}

// @router /sources/streams[post]
func (c *SourceHandler) GetSourceCatalog() {
	var req models.CreateSourceRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}
	var catalog map[string]interface{}
	var err error
	// Try to use Temporal if available
	if c.tempClient != nil {
		fmt.Println("Using Temporal workflow for catalog discovery")

		// Create a unique workflow ID
		// workflowID := fmt.Sprintf("discover-catalog-%s-%d", req.Type, time.Now().Unix())
		// fmt.Printf("Starting workflow with ID: %s\n", workflowID)
		// Execute the workflow using Temporal
		catalog, err = c.tempClient.GetCatalog(
			c.Ctx.Request.Context(),
			req.Type,
			req.Version,
			req.Config,
		)
	}
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to get catalog: %v", err))
		return
	}
	utils.SuccessResponse(&c.Controller, catalog)
}

// @router /sources/:id/jobs [get]
func (c *SourceHandler) GetSourceJobs() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	// Check if source exists
	_, err = c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Create a job ORM and get jobs by source ID
	jobORM := database.NewJobORM()
	jobs, err := jobORM.GetBySourceID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs")
		return
	}

	// Format as required by API contract
	response := map[string]interface{}{
		"jobs": jobs,
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/sources/versions [get]
func (c *SourceHandler) GetSourceVersions() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	_, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Get source type from query parameter
	sourceType := c.GetString("type")
	if sourceType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Source type is required")
		return
	}

	// In a real implementation, we would query for available versions
	// based on the source type and project ID
	// For now, we'll return a mock response

	// Mock available versions (this would be replaced with actual DB query)
	versions := []string{"latest"}

	response := map[string]interface{}{
		"version": versions,
	}

	utils.SuccessResponse(&c.Controller, response)
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
					"description": "Hostname or IP address of the PostgreSQL server",
					"order":       1,
				},
				"port": map[string]interface{}{
					"type":        "integer",
					"title":       "Postgres Port",
					"description": "Port number of the PostgreSQL server",
					"default":     5432,
					"order":       2,
				},
				"database": map[string]interface{}{
					"type":        "string",
					"title":       "Database Name",
					"description": "Name of the PostgreSQL database",
					"order":       3,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "Database user for authentication",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "Password for the database user",
					"format":      "password",
					"order":       5,
				},
				"jdbc_url_params": map[string]interface{}{
					"type":        "string",
					"title":       "JDBC URL Parameters",
					"description": "Optional JDBC parameters as key-value pairs",
					"order":       6,
				},
				"ssl": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"mode": map[string]interface{}{
							"type":        "string",
							"title":       "SSL Mode",
							"description": "SSL mode to connect (disable, require, verify-ca, etc.)",
							"default":     "disable",
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
							"description": "Slot name for CDC",
							"default":     "postgres_slot",
						},
						"intial_wait_time": map[string]interface{}{
							"type":        "integer",
							"title":       "Initial Wait Time",
							"description": "Seconds to wait before starting CDC",
							"default":     10,
						},
					},
					"order": 8,
				},
				"reader_batch_size": map[string]interface{}{
					"type":        "integer",
					"title":       "Reader Batch Size",
					"description": "Number of records to read in each batch",
					"default":     100000,
					"order":       9,
				},
				"default_mode": map[string]interface{}{
					"type":        "string",
					"title":       "Default Mode",
					"description": "Extraction mode (e.g., full or cdc)",
					"default":     "cdc",
					"order":       10,
				},
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Number of threads to use for backfill",
					"default":     5,
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
					"description": "Comma-separated list of MySQL hosts",
					"default":     "mysql-host",
					"order":       1,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "MySQL username",
					"default":     "mysql-user",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "Password for the MySQL user",
					"format":      "password",
					"default":     "mysql-password",
					"order":       5,
				},
				"database": map[string]interface{}{
					"type":        "string",
					"title":       "Database",
					"description": "Target MySQL database name",
					"default":     "mysql-database",
					"order":       3,
				},
				"port": map[string]interface{}{
					"type":        "integer",
					"title":       "Port",
					"description": "Port number for MySQL",
					"default":     3306,
					"order":       2,
				},
				"update_method": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"intial_wait_time": map[string]interface{}{
							"type":        "integer",
							"title":       "Initial Wait Time",
							"description": "Wait time in seconds before retrying",
							"default":     10,
						},
					},
					"order": 6,
				},
				"tls_skip_verify": map[string]interface{}{
					"type":        "boolean",
					"title":       "Skip TLS Verification",
					"description": "Whether to skip TLS certificate verification",
					"default":     true,
					"order":       10,
				},
				"default_mode": map[string]interface{}{
					"type":        "string",
					"title":       "Default Mode",
					"description": "Extraction mode (e.g., full or cdc)",
					"default":     "cdc",
					"order":       7,
				},
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Number of parallel threads",
					"default":     5,
					"order":       8,
				},
				"backoff_retry_count": map[string]interface{}{
					"type":        "integer",
					"title":       "Backoff Retry Count",
					"description": "Retry attempts before failing",
					"default":     2,
					"order":       9,
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
					"description": "List of MongoDB hosts (with port)",
					"items":       map[string]interface{}{"type": "string"},
					"default":     []string{"host1:27017", "host2:27017"},
					"order":       1,
				},
				"username": map[string]interface{}{
					"type":        "string",
					"title":       "Username",
					"description": "MongoDB username",
					"default":     "test",
					"order":       4,
				},
				"password": map[string]interface{}{
					"type":        "string",
					"title":       "Password",
					"description": "MongoDB password",
					"format":      "password",
					"default":     "test",
					"order":       5,
				},
				"authdb": map[string]interface{}{
					"type":        "string",
					"title":       "Auth DB",
					"description": "Authentication database",
					"default":     "admin",
					"order":       3,
				},
				"replica-set": map[string]interface{}{
					"type":        "string",
					"title":       "Replica Set",
					"description": "MongoDB replica set name",
					"default":     "rs0",
					"order":       6,
				},
				"read-preference": map[string]interface{}{
					"type":        "string",
					"title":       "Read Preference",
					"description": "Read preference (e.g., primary, secondaryPreferred)",
					"default":     "",
					"order":       7,
				},
				"srv": map[string]interface{}{
					"type":        "boolean",
					"title":       "Use SRV",
					"description": "Whether to use DNS SRV",
					"default":     false,
					"order":       8,
				},
				"server-ram": map[string]interface{}{
					"type":        "integer",
					"title":       "Server RAM",
					"description": "Server memory in GB",
					"default":     16,
					"order":       13,
				},
				"database": map[string]interface{}{
					"type":        "string",
					"title":       "Database Name",
					"description": "MongoDB target database",
					"default":     "database",
					"order":       2,
				},
				"max_threads": map[string]interface{}{
					"type":        "integer",
					"title":       "Max Threads",
					"description": "Maximum threads to use for ingestion",
					"default":     5,
					"order":       9,
				},
				"default_mode": map[string]interface{}{
					"type":        "string",
					"title":       "Default Mode",
					"description": "Extraction mode (e.g., full_refresh, cdc)",
					"default":     "cdc",
					"order":       10,
				},
				"backoff_retry_count": map[string]interface{}{
					"type":        "integer",
					"title":       "Retry Count",
					"description": "Number of retries before failure",
					"default":     2,
					"order":       11,
				},
				"partition_strategy": map[string]interface{}{
					"type":        "string",
					"title":       "Partition Strategy",
					"description": "Strategy for collection partitioning",
					"default":     "",
					"order":       12,
				},
			},
			"required": []string{"hosts", "username", "password", "database"},
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
