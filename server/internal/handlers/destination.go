package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"go.temporal.io/api/workflowservice/v1"

	"github.com/datazip/olake-server/internal/constants"
	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/internal/temporal"
	"github.com/datazip/olake-server/utils"
)

type DestHandler struct {
	web.Controller
	destORM    *database.DestinationORM
	jobORM     *database.JobORM
	tempClient *temporal.Client
}

func (c *DestHandler) Prepare() {
	c.destORM = database.NewDestinationORM()
	c.jobORM = database.NewJobORM()
	tempAddress := web.AppConfig.DefaultString("TEMPORAL_ADDRESS", "localhost:7233")
	tempClient, err := temporal.NewClient(tempAddress)
	if err != nil {
		// Log the error but continue - we'll fall back to direct Docker execution if Temporal fails
		logs.Error("Failed to create Temporal client: %v", err)
	} else {
		c.tempClient = tempClient
	}
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	// Get project ID from path
	//use olake project id when is needed
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	destinations, err := c.destORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve destinations")
		return
	}

	// Filter destinations by project ID in memory
	// In a real implementation, this would be done in the database query
	// var filteredDestinations []*models.Destination
	// for _, dest := range destinations {
	// 	if dest.ProjectID == projectIDStr {
	// 		filteredDestinations = append(filteredDestinations, dest)
	// 	}
	// }

	// Format response data
	destItems := make([]models.DestinationDataItem, 0, len(destinations))
	for _, dest := range destinations {
		item := models.DestinationDataItem{
			ID:        dest.ID,
			Name:      dest.Name,
			Type:      dest.DestType,
			Version:   dest.Version,
			Config:    dest.Config,
			CreatedAt: dest.CreatedAt.Format(time.RFC3339),
			UpdatedAt: dest.UpdatedAt.Format(time.RFC3339),
		}

		// Add creator username if available
		if dest.CreatedBy != nil {
			item.CreatedBy = dest.CreatedBy.Username
		}

		// Add updater username if available
		if dest.UpdatedBy != nil {
			item.UpdatedBy = dest.UpdatedBy.Username
		}
		jobs, err := c.jobORM.GetByDestinationID(dest.ID)
		destJobs := make([]models.JobDataItem, 0, len(jobs))
		if err == nil {
			for _, job := range jobs {
				jobInfo := models.JobDataItem{
					Name:     job.Name,
					ID:       job.ID,
					Activate: job.Active,
				}

				// Add destination name if available
				if job.DestID != nil {
					jobInfo.SourceName = job.SourceID.Name
					jobInfo.SourceType = job.SourceID.Type
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
				destJobs = append(destJobs, jobInfo)
			}

		}
		item.Jobs = destJobs

		destItems = append(destItems, item)

	}
	utils.SuccessResponse(&c.Controller, destItems)
}

// @router /project/:projectid/destinations [post]
func (c *DestHandler) CreateDestination() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")

	var req models.CreateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert request to Destination model
	destination := &models.Destination{
		Name:      req.Name,
		DestType:  req.Type,
		Version:   req.Version,
		Config:    req.Config,
		ProjectID: projectIDStr,
	}

	// Set created by if user is logged in
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		destination.CreatedBy = user
		destination.UpdatedBy = user
	}

	if err := c.destORM.Create(destination); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create destination: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		fmt.Println("Error converting project ID to int:", err)
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Project ID will be used for permission checking in future implementations
	_ = projectID

	// Get destination ID from path
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	var req models.UpdateDestinationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get existing destination
	existingDest, err := c.destORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}

	// Update fields
	existingDest.Name = req.Name
	existingDest.DestType = req.Type
	existingDest.Version = req.Version
	existingDest.Config = req.Config
	existingDest.UpdatedAt = time.Now()

	// Update user who made changes
	userID := c.GetSession(constants.SessionUserID)
	if userID != nil {
		user := &models.User{ID: userID.(int)}
		existingDest.UpdatedBy = user
	}

	if err := c.destORM.Update(existingDest); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update destination")
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		fmt.Println("Error converting project ID to int:", err)
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Project ID will be used for permission checking in future implementations
	_ = projectID

	// Get destination ID from path
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	// Get the name for the response
	name, err := c.destORM.GetName(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}
	jobs, err := c.jobORM.GetBySourceID(id)
	for _, job := range jobs {
		job.Active = false
	}
	if err := c.destORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete destination")
		return
	}

	response := models.DeleteDestinationResponse{
		Name: name,
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	// Get project ID from path (not used in current implementation)
	_ = c.Ctx.Input.Param(":projectid")
	// Will be used for multi-tenant filtering in the future
	var req models.DestinationTestConnectionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Type == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	if req.Version == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination version is required")
		return
	}

	// For now, always return success
	utils.SuccessResponse(&c.Controller, req)
}

// @router /destinations/:id/jobs [get]
func (c *DestHandler) GetDestinationJobs() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid destination ID")
		return
	}

	// Check if destination exists
	_, err = c.destORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "Destination not found")
		return
	}

	// Create a job ORM and get jobs by destination ID
	jobORM := database.NewJobORM()
	jobs, err := jobORM.GetByDestinationID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve jobs")
		return
	}

	// Format as required by API contract
	response := map[string]interface{}{
		"jobs": jobs,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	// Get project ID from path
	projectIDStr := c.Ctx.Input.Param(":projectid")
	_, err := strconv.Atoi(projectIDStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Get destination type from query parameter
	destType := c.GetString("type")
	if destType == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	// In a real implementation, we would query for available versions
	// based on the destination type and project ID
	// For now, we'll return a mock response

	// Mock available versions (this would be replaced with actual DB query)
	versions := []string{"latest"}

	response := map[string]interface{}{
		"version": versions,
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/destinations/spec [post]
func (c *DestHandler) GetDestinationSpec() {
	// Get project ID from path (not used in current implementation)
	_ = c.Ctx.Input.Param(":projectid")
	// Will be used for multi-tenant filtering in the future

	var req models.SpecRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Type == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination type is required")
		return
	}

	if req.Version == "" {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Destination version is required")
		return
	}

	var spec interface{}
	var uiSchema interface{}

	switch req.Type {
	case "s3":
		spec = map[string]interface{}{
			"title": "Writer Settings",
			"type":  "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"title":       "File Type",
					"description": "Type of file to write (e.g., PARQUET)",
					"enum":        []string{"PARQUET"},
					"default":     "PARQUET",
				},
				"writer": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"normalization": map[string]interface{}{
							"type":        "boolean",
							"title":       "Normalization",
							"description": "Whether to normalize the data before writing",
							"default":     false,
						},
						"s3_bucket": map[string]interface{}{
							"type":        "string",
							"title":       "S3 Bucket",
							"description": "Name of the S3 bucket",
						},
						"s3_region": map[string]interface{}{
							"type":        "string",
							"title":       "S3 Region",
							"description": "AWS region where the bucket is located",
						},
						"s3_access_key": map[string]interface{}{
							"type":        "string",
							"title":       "AWS Access Key",
							"description": "AWS access key ID",
							"format":      "password",
						},
						"s3_secret_key": map[string]interface{}{
							"type":        "string",
							"title":       "AWS Secret Key",
							"description": "AWS secret access key",
							"format":      "password",
						},
						"s3_path": map[string]interface{}{
							"type":        "string",
							"title":       "S3 Path",
							"description": "Path within the S3 bucket where files will be written",
							"default":     "/",
						},
					},
					"required": []string{"s3_bucket", "s3_region", "s3_access_key", "s3_secret_key"},
				},
			},
			"required": []string{"type", "writer"},
		}
		uiSchema = map[string]interface{}{
			"type": map[string]interface{}{
				"ui:widget": "hidden",
			},
		}
	case "iceberg":
		// Get catalog type from request
		var catalogSpec interface{}
		var catalogUiSchema interface{}

		switch req.Catalog {
		case "glue":
			catalogSpec = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"catalog_type": map[string]interface{}{
						"type":        "string",
						"title":       "Catalog Type",
						"description": "Type of catalog to use",
						"enum":        []string{"glue"},
						"default":     "glue",
					},
					"normalization": map[string]interface{}{
						"type":        "boolean",
						"title":       "Normalization",
						"description": "Whether to normalize the data before writing",
						"default":     false,
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path",
						"description": "S3 path for Iceberg tables",
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "AWS region for Glue catalog",
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "AWS access key ID",
						"format":      "password",
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "AWS secret access key",
						"format":      "password",
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Name of the Iceberg database",
					},
					"grpc_port": map[string]interface{}{
						"type":        "integer",
						"title":       "gRPC Port",
						"description": "Port for gRPC communication",
						"default":     50051,
					},
					"server_host": map[string]interface{}{
						"type":        "string",
						"title":       "Server Host",
						"description": "Host for server communication",
						"default":     "localhost",
					},
				},
				"required": []string{"catalog_type", "normalization", "iceberg_s3_path", "aws_region", "aws_access_key", "aws_secret_key", "iceberg_db"},
			}
			catalogUiSchema = map[string]interface{}{
				"catalog_type": map[string]interface{}{
					"ui:widget": "hidden",
				},
			}

		case "rest":
			catalogSpec = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"catalog_type": map[string]interface{}{
						"type":        "string",
						"title":       "Catalog Type",
						"description": "Type of catalog to use",
						"enum":        []string{"rest"},
						"default":     "rest",
					},
					"normalization": map[string]interface{}{
						"type":        "boolean",
						"title":       "Normalization",
						"description": "Whether to normalize the data before writing",
						"default":     false,
					},
					"rest_catalog_url": map[string]interface{}{
						"type":        "string",
						"title":       "REST Catalog URL",
						"description": "URL for REST catalog service",
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path",
						"description": "S3 path for Iceberg tables",
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Name of the Iceberg database",
					},
				},
				"required": []string{"catalog_type", "normalization", "rest_catalog_url", "iceberg_s3_path", "iceberg_db"},
			}
			catalogUiSchema = map[string]interface{}{
				"catalog_type": map[string]interface{}{
					"ui:widget": "hidden",
				},
			}

		case "jdbc":
			catalogSpec = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"catalog_type": map[string]interface{}{
						"type":        "string",
						"title":       "Catalog Type",
						"description": "Type of catalog to use",
						"enum":        []string{"jdbc"},
						"default":     "jdbc",
					},
					"normalization": map[string]interface{}{
						"type":        "boolean",
						"title":       "Normalization",
						"description": "Whether to normalize the data before writing",
						"default":     false,
					},
					"jdbc_url": map[string]interface{}{
						"type":        "string",
						"title":       "JDBC URL",
						"description": "JDBC connection URL",
					},
					"jdbc_username": map[string]interface{}{
						"type":        "string",
						"title":       "JDBC Username",
						"description": "JDBC connection username",
					},
					"jdbc_password": map[string]interface{}{
						"type":        "string",
						"title":       "JDBC Password",
						"description": "JDBC connection password",
						"format":      "password",
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path",
						"description": "S3 path for Iceberg tables",
					},
					"s3_endpoint": map[string]interface{}{
						"type":        "string",
						"title":       "S3 Endpoint",
						"description": "S3 endpoint URL",
					},
					"s3_use_ssl": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use SSL for S3",
						"description": "Whether to use SSL for S3 connections",
						"default":     false,
					},
					"s3_path_style": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use Path Style for S3",
						"description": "Whether to use path style for S3 URLs",
						"default":     true,
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "AWS access key ID",
						"format":      "password",
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "AWS region for S3",
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "AWS secret access key",
						"format":      "password",
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Name of the Iceberg database",
					},
				},
				"required": []string{"catalog_type", "jdbc_url", "jdbc_username", "jdbc_password", "iceberg_s3_path", "aws_access_key", "aws_secret_key", "aws_region", "iceberg_db"},
			}
			catalogUiSchema = map[string]interface{}{
				"catalog_type": map[string]interface{}{
					"ui:widget": "hidden",
				},
			}

		case "hive":
			catalogSpec = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"catalog_type": map[string]interface{}{
						"type":        "string",
						"title":       "Catalog Type",
						"description": "Type of catalog to use",
						"enum":        []string{"hive"},
						"default":     "hive",
					},
					"normalization": map[string]interface{}{
						"type":        "boolean",
						"title":       "Normalization",
						"description": "Whether to normalize the data before writing",
						"default":     false,
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path",
						"description": "S3 path for Iceberg tables",
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "AWS region for S3",
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "AWS access key ID",
						"format":      "password",
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "AWS secret access key",
						"format":      "password",
					},
					"s3_endpoint": map[string]interface{}{
						"type":        "string",
						"title":       "S3 Endpoint",
						"description": "S3 endpoint URL",
					},
					"hive_uri": map[string]interface{}{
						"type":        "string",
						"title":       "Hive URI",
						"description": "URI for Hive metastore",
					},
					"s3_use_ssl": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use SSL for S3",
						"description": "Whether to use SSL for S3 connections",
						"default":     false,
					},
					"s3_path_style": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use Path Style for S3",
						"description": "Whether to use path style for S3 URLs",
						"default":     true,
					},
					"hive_clients": map[string]interface{}{
						"type":        "integer",
						"title":       "Hive Clients",
						"description": "Number of Hive clients",
						"default":     5,
					},
					"hive_sasl_enabled": map[string]interface{}{
						"type":        "boolean",
						"title":       "Enable SASL for Hive",
						"description": "Whether to enable SASL for Hive connections",
						"default":     false,
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Name of the Iceberg database",
					},
				},
				"required": []string{"catalog_type", "iceberg_s3_path", "aws_region", "aws_access_key", "aws_secret_key", "iceberg_db"},
			}
			catalogUiSchema = map[string]interface{}{
				"catalog_type": map[string]interface{}{
					"ui:widget": "hidden",
				},
			}

		default:
			utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Unsupported catalog type")
			return
		}

		spec = map[string]interface{}{
			"title": "Writer Settings",
			"type":  "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"title":       "File Type",
					"description": "Type of file to write",
					"enum":        []string{"ICEBERG"},
					"default":     "ICEBERG",
				},
				"writer": catalogSpec,
			},
			"required": []string{"type", "writer"},
		}

		uiSchema = map[string]interface{}{
			"type": map[string]interface{}{
				"ui:widget": "hidden",
			},
			"writer": catalogUiSchema,
		}

	default:
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Unsupported destination type")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Form initialized successfully",
		"data": map[string]interface{}{
			"version":  req.Version,
			"type":     req.Type,
			"spec":     spec,
			"uiSchema": uiSchema,
		},
	}

	c.Data["json"] = response
	c.ServeJSON()
}
