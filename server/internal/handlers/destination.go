package handlers

import (
	"context"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/services"
	"github.com/datazip/olake-frontend/server/utils"
)

type DestHandler struct {
	web.Controller
	destService *services.DestinationService
}

func (c *DestHandler) Prepare() {
	var err error
	c.destService, err = services.NewDestinationService()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to initialize destination service")
		return
	}
}

// @router /project/:projectid/destinations [get]
func (c *DestHandler) GetAllDestinations() {
	projectID := c.Ctx.Input.Param(":projectid")

	destinations, err := c.destService.GetAllDestinations(context.Background(), projectID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, destinations)
}

// @router /project/:projectid/destinations [post]
func (c *DestHandler) CreateDestination() {
	projectID := c.Ctx.Input.Param(":projectid")

	var req models.CreateDestinationRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)

	if err := c.destService.CreateDestination(context.Background(), req, projectID, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [put]
func (c *DestHandler) UpdateDestination() {
	id := GetIDFromPath(&c.Controller)

	var req models.UpdateDestinationRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID := GetUserIDFromSession(&c.Controller)

	if err := c.destService.UpdateDestination(context.Background(), id, req, userID); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /project/:projectid/destinations/:id [delete]
func (c *DestHandler) DeleteDestination() {
	id := GetIDFromPath(&c.Controller)

	response, err := c.destService.DeleteDestination(context.Background(), id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, response)
}

// @router /project/:projectid/destinations/test [post]
func (c *DestHandler) TestConnection() {
	var req models.DestinationTestConnectionRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	result, err := c.destService.TestConnection(context.Background(), req)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(&c.Controller, result)
}

// @router /destinations/:id/jobs [get]
func (c *DestHandler) GetDestinationJobs() {
	id := GetIDFromPath(&c.Controller)

	jobs, err := c.destService.GetDestinationJobs(context.Background(), id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"jobs": jobs,
	})
}

// @router /project/:projectid/destinations/versions [get]
func (c *DestHandler) GetDestinationVersions() {
	destType := c.GetString("type")

	versions, err := c.destService.GetDestinationVersions(context.Background(), destType)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version": versions,
	})
}

// @router /project/:projectid/destinations/spec [post]
func (c *DestHandler) GetDestinationSpec() {
	// Get project ID from path (not used in current implementation)
	_ = c.Ctx.Input.Param(":projectid")
	// Will be used for multi-tenant filtering in the future

	var req models.SpecRequest
	if err := bindJSON(&c.Controller, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
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
						"s3_bucket": map[string]interface{}{
							"type":        "string",
							"title":       "S3 Bucket",
							"description": "The name of an existing Amazon S3 bucket with appropriate access permissions to store output files",
							"order":       1,
						},
						"s3_region": map[string]interface{}{
							"type":        "string",
							"title":       "S3 Region",
							"description": "Specify the AWS region where the S3 bucket is hosted",
							"order":       2,
						},
						"s3_access_key": map[string]interface{}{
							"type":        "string",
							"title":       "AWS Access Key",
							"description": "The AWS access key for authenticating S3 requests, typically a 20-character alphanumeric string",
							"format":      "password",
							"order":       3,
						},
						"s3_secret_key": map[string]interface{}{
							"type":        "string",
							"title":       "AWS Secret Key",
							"description": "The AWS secret key for S3 authentication—typically 40+ characters long",
							"format":      "password",
							"order":       4,
						},
						"s3_path": map[string]interface{}{
							"type":        "string",
							"title":       "S3 Path",
							"description": "Specify the S3 bucket path (prefix) where data files will be written, typically starting with a '/' (e.g., '/data')",
							"order":       5,
							"default":     "",
						},
					},
					"required": []string{"s3_bucket", "s3_region"},
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
		var catalogUISchema interface{}

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
						"order":       1,
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path (Warehouse)",
						"description": "Specifies the S3 path in AWS where Iceberg data is stored",
						"order":       2,
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "Specify the AWS region where the S3 bucket is hosted",
						"order":       3,
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "The AWS access key for authenticating S3 requests, typically a 20-character alphanumeric string",
						"format":      "password",
						"order":       4,
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "The AWS secret key for S3 authentication—typically 40+ characters long",
						"format":      "password",
						"order":       5,
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Specifies the name of the database to be created in AWS Glue",
						"order":       6,
					},
					"grpc_port": map[string]interface{}{
						"type":        "integer",
						"title":       "gRPC Port",
						"description": "Port on which the gRPC server listens (mostly 50051)",
						"default":     50051,
						"order":       7,
					},
					"server_host": map[string]interface{}{
						"type":        "string",
						"title":       "Server Host",
						"description": "Host address of the gRPC server",
						"default":     "localhost",
						"order":       8,
					},
				},
				"required": []string{"catalog_type", "iceberg_s3_path", "aws_region", "iceberg_db"},
			}
			catalogUISchema = map[string]interface{}{
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
						"order":       1,
					},
					"rest_catalog_url": map[string]interface{}{
						"type":        "string",
						"title":       "REST Catalog URI",
						"description": "Specifies the endpoint URI for the REST catalog service the writer will connect to",
						"order":       2,
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path (Warehouse)",
						"description": "The S3 path or storage location for Iceberg data",
						"order":       3,
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Specifies the name of the Iceberg database that will be used by the destination",
						"order":       4,
					},
					"token": map[string]interface{}{
						"type":        "string",
						"title":       "Token",
						"description": "Optional token used for authenticating with the REST catalog",
						"format":      "password",
						"order":       6,
					},
					"rest_auth_type": map[string]interface{}{
						"type":        "string",
						"title":       "REST Auth Type",
						"description": "Type of authentication to use with the REST catalog. Need to use it for only `oauth2`",
						"order":       7,
					},
					"oauth2_uri": map[string]interface{}{
						"type":        "string",
						"title":       "OAuth2 Auth URI",
						"description": "OAuth2 authorization URI for obtaining access tokens",
						"order":       8,
					},

					"scope": map[string]interface{}{
						"type":        "string",
						"title":       "Scope (Oauth2)",
						"description": "OAuth2 scope to be used during token acquisition",
						"order":       9,
					},
					"credential": map[string]interface{}{
						"type":        "string",
						"title":       "Credential (Oauth2)",
						"description": "Optional credential used for authenticating REST requests",
						"format":      "password",
						"order":       10,
					},
					"no_identifier_fields": map[string]interface{}{
						"type":        "boolean",
						"title":       "Disable Identifier Tables",
						"description": "Disable creation of Iceberg identifier tables for this catalog, Needed for environments which doesn't support equality deletes based updates (Ex, Databricks unity managed Iceberg tables)",
						"order":       11,
					},
					"s3_endpoint": map[string]interface{}{
						"type":        "string",
						"title":       "S3 Endpoint",
						"description": "Specifies the endpoint URL for the S3 service (e.g., MinIO)",
						"order":       12,
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "Specify the AWS region where the S3 bucket is hosted",
						"order":       13,
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "The AWS access key for authenticating S3 requests, typically a 20-character alphanumeric string",
						"format":      "password",
						"order":       14,
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "The AWS secret key for S3 authentication—typically 40+ characters long",
						"format":      "password",
						"order":       15,
					},
					"rest_signing_name": map[string]interface{}{
						"type":        "string",
						"title":       "REST Signing Name",
						"description": "Optional AWS signing name to be used when authenticating REST requests",
						"order":       16,
					},
					"rest_signing_region": map[string]interface{}{
						"type":        "string",
						"title":       "REST Signing Region",
						"description": "AWS region used for signing REST requests",
						"order":       16,
					},
					"rest_signing_v_4": map[string]interface{}{
						"type":        "boolean",
						"title":       "Rest Enable Signature V4",
						"description": "Enable AWS Signature Version 4 for REST request signing",
						"order":       17,
					},
				},
				"required": []string{"catalog_type", "rest_catalog_url", "iceberg_s3_path", "iceberg_db"},
			}
			catalogUISchema = map[string]interface{}{
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
						"order":       1,
					},
					"jdbc_url": map[string]interface{}{
						"type":        "string",
						"title":       "JDBC URL",
						"description": "Specifies the JDBC URL used to connect to the PostgreSQL catalog",
						"order":       2,
					},
					"jdbc_username": map[string]interface{}{
						"type":        "string",
						"title":       "JDBC Username",
						"description": "Specifies the username used to authenticate with the PostgreSQL catalog",
						"order":       3,
					},
					"jdbc_password": map[string]interface{}{
						"type":        "string",
						"title":       "JDBC Password",
						"description": "Specifies the password used to authenticate with the PostgreSQL catalog",
						"format":      "password",
						"order":       4,
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path (Warehouse)",
						"description": "Specifies the S3 path for storing Iceberg data and metadata files",
						"order":       5,
					},
					"s3_endpoint": map[string]interface{}{
						"type":        "string",
						"title":       "S3 Endpoint",
						"description": "Specifies the endpoint URL for the S3 service (e.g., MinIO)",
						"order":       6,
					},
					"s3_use_ssl": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use SSL for S3",
						"description": "Boolean flag indicating whether to use SSL when connecting to S3",
						"default":     false,
						"order":       7,
					},
					"s3_path_style": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use Path Style for S3",
						"description": "Enables path-style addressing for S3 API requests",
						"default":     true,
						"order":       8,
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "The AWS access key for authenticating S3 requests, typically a 20-character alphanumeric string",
						"format":      "password",
						"order":       9,
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "Specify the AWS region where the S3 bucket is hosted",
						"order":       10,
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "The AWS secret key for S3 authentication—typically 40+ characters long",
						"format":      "password",
						"order":       11,
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Specifies the name of the Iceberg database to be used in the destination",
						"order":       12,
					},
				},
				"required": []string{"catalog_type", "jdbc_url", "jdbc_username", "jdbc_password", "iceberg_s3_path", "aws_region", "iceberg_db"},
			}
			catalogUISchema = map[string]interface{}{
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
						"order":       1,
					},
					"iceberg_s3_path": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg S3 Path (Warehouse)",
						"description": "Specifies the S3 path for storing Iceberg data, such as 's3a://warehouse/', representing the target bucket or directory",
						"order":       2,
					},
					"aws_region": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Region",
						"description": "Specify the AWS region where the S3 bucket is hosted",
						"order":       3,
					},
					"aws_access_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Access Key",
						"description": "The AWS access key for authenticating S3 requests, typically a 20-character alphanumeric string",
						"format":      "password",
						"order":       4,
					},
					"aws_secret_key": map[string]interface{}{
						"type":        "string",
						"title":       "AWS Secret Key",
						"description": "The AWS secret key for S3 authentication—typically 40+ characters long",
						"format":      "password",
						"order":       5,
					},
					"s3_endpoint": map[string]interface{}{
						"type":        "string",
						"title":       "S3 Endpoint",
						"description": "Specifies the S3 service endpoint URL, used when connecting to S3-compatible storage like MinIO (e.g., on localhost)",
						"order":       6,
					},
					"hive_uri": map[string]interface{}{
						"type":        "string",
						"title":       "Hive URI",
						"description": "Specifies the URI of the Hive Metastore service used for catalog interactions by the writer",
						"order":       7,
					},
					"s3_use_ssl": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use SSL for S3",
						"description": "Indicates if SSL is enabled for S3 connections; 'false' disables SSL",
						"order":       8,
					},
					"s3_path_style": map[string]interface{}{
						"type":        "boolean",
						"title":       "Use Path Style for S3",
						"description": "Specifies whether path-style access is used for S3; 'true' enables path-style addressing instead of the default virtual-hosted style",
						"order":       9,
					},
					"hive_clients": map[string]interface{}{
						"type":        "integer",
						"title":       "Hive Clients",
						"description": "Specifies the number of Hive clients allocated to handle interactions with the Hive Metastore",
						"order":       10,
					},
					"hive_sasl_enabled": map[string]interface{}{
						"type":        "boolean",
						"title":       "Enable SASL for Hive",
						"description": "Indicates if SASL authentication is enabled for the Hive connection; 'false' means SASL is disabled",
						"order":       11,
					},
					"iceberg_db": map[string]interface{}{
						"type":        "string",
						"title":       "Iceberg Database",
						"description": "Specifies the name of the Iceberg database to be used in the destination",
						"order":       12,
					},
				},
				"required": []string{"catalog_type", "iceberg_s3_path", "aws_region", "iceberg_db"},
			}
			catalogUISchema = map[string]interface{}{
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
			"writer": catalogUISchema,
		}

	default:
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Unsupported destination type")
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"version":  req.Version,
		"type":     req.Type,
		"spec":     spec,
		"uiSchema": uiSchema,
	})
}
