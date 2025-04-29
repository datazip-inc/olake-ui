package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/docker"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/utils"
)

type SourceHandler struct {
	web.Controller
	sourceORM *database.SourceORM
	userORM   *database.UserORM
	jobORM    *database.JobORM
}

func (c *SourceHandler) Prepare() {
	c.sourceORM = database.NewSourceORM()
	c.userORM = database.NewUserORM()
	c.jobORM = database.NewJobORM()
}

// @router /project/:projectid/sources [get]
func (c *SourceHandler) GetAllSources() {
	sources, err := c.sourceORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve sources")
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
		sourceJobs := make([]map[string]interface{}, 0) // always initialize
		if err == nil {
			for _, job := range jobs {
				jobInfo := map[string]interface{}{
					"name":     job.Name,
					"id":       job.ID,
					"activate": job.Active,
				}
				// Add destination name if available
				if job.DestID != nil {
					jobInfo["dest_name"] = job.DestID.Name
					jobInfo["dest_type"] = job.DestID.DestType
				}

				// Add hardcoded last run info (or parse from job.State if needed)
				jobInfo["last_run_time"] = "2025-04-27T15:30:00Z"
				jobInfo["last_run_state"] = "success"

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

	// For now, always return success
	utils.SuccessResponse(&c.Controller, req)

}

// @router /sources/:id/catalog [get]
func (c *SourceHandler) GetSourceCatalog() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid source ID")
		return
	}

	// Get existing source
	source, err := c.sourceORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Source not found")
		return
	}

	// Initialize Docker runner
	runner := docker.NewRunner(docker.GetDefaultConfigDir())

	// Execute Docker command to get catalog
	catalog, err := runner.GetCatalog(source.Type, source.Version, source.Config, source.ID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to generate catalog: %v", err))
		return
	}

	// Return catalog data
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
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}

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

	var spec string

	switch req.Type {
	case "postgres":
		spec = `{ "host": "string", "port": "integer", "database": "string", "username": "string", "password": "string", "jdbc_url_params": "object", "ssl": { "mode": "string" }, "update_method": { "replication_slot": "string", "intial_wait_time": "integer" }, "reader_batch_size": "integer", "default_mode": "string", "max_threads": "integer" }`
	case "mysql":
		spec = `{ "hosts": "string", "username": "string", "password": "string", "database": "string", "port": "integer", "update_method": { "intial_wait_time": "integer" }, "tls_skip_verify": "boolean", "default_mode": "string", "max_threads": "integer", "backoff_retry_count": "integer" }`
	case "mongodb":
		spec = `{ "hosts": ["string"], "username": "string", "password": "string", "authdb": "string", "replica-set": "string", "read-preference": "string", "srv": "boolean", "server-ram": "integer", "database": "string", "max_threads": "integer", "default_mode": "string", "backoff_retry_count": "integer", "partition_strategy": "string" }`
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
