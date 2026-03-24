package optimisation

import (
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	cmpModels "github.com/datazip-inc/olake-ui/server/internal/services/optimisation/models"
	optModels "github.com/datazip-inc/olake-ui/server/internal/services/optimisation/models"
	"github.com/datazip-inc/olake-ui/server/utils"
)

func MapoptimisationCatalogToOLakeConfig(catalog *optModels.CatalogRequest) (*models.Config, error) {
	config := &models.Config{}

	// Map catalog name
	config.CatalogName = catalog.Name

	// Map catalog type back to OLake format
	config.CatalogType = mapoptimisationTypeToOLake(catalog.Type)

	// Map storage and auth config
	if catalog.StorageConfig != nil {
		config.Region = catalog.StorageConfig["storage.s3.region"]
		config.S3Endpoint = catalog.StorageConfig["storage.s3.endpoint"]
	}

	if catalog.AuthConfig != nil {
		config.AccessKey = catalog.AuthConfig["auth.ak_sk.access_key"]
		config.SecretKey = catalog.AuthConfig["auth.ak_sk.secret_key"]
	}

	// Map properties based on catalog type
	if catalog.Properties != nil {
		config.IcebergS3Path = catalog.Properties["warehouse"]

		// Map S3 properties from catalog properties (for JDBC catalogs)
		if catalog.Properties["endpoint"] != "" {
			config.S3Endpoint = catalog.Properties["endpoint"]
		}
		if catalog.Properties["access-key-id"] != "" {
			config.AccessKey = catalog.Properties["access-key-id"]
		}
		if catalog.Properties["secret-access-key"] != "" {
			config.SecretKey = catalog.Properties["secret-access-key"]
		}

		switch strings.ToLower(string(config.CatalogType)) {
		case "glue":
			config.GlueCatalogID = catalog.Properties["glue.id"]
			config.GlueEndpoint = catalog.Properties["glue.endpoint"]
			config.GlueRegion = catalog.Properties["client.region"]
			if config.GlueCatalogID != "" || config.GlueEndpoint != "" {
				config.UseGlueAdditionalConfig = true
			}

		case "jdbc":
			config.JDBCUrl = catalog.Properties["uri"]
			config.JDBCUsername = catalog.Properties["jdbc.user"]
			config.JDBCPassword = catalog.Properties["jdbc.password"]

		case "hive":
			config.HiveURI = catalog.Properties["uri"]
			if clientsStr := catalog.Properties["clients"]; clientsStr != "" {
				fmt.Sscanf(clientsStr, "%d", &config.HiveClients)
			}

		case "rest":
			config.RestCatalogURL = catalog.Properties["uri"]
			config.RestToken = catalog.Properties["token"]
			config.RestOAuthURI = catalog.Properties["oauth2-server-uri"]
			config.RestAuthType = catalog.Properties["rest.auth.type"]
			config.RestCredential = catalog.Properties["credential"]
			config.RestScope = catalog.Properties["scope"]
			if catalog.Properties["rest.sigv4-enabled"] == "true" {
				config.RestSigningV4 = true
				config.RestSigningName = catalog.Properties["rest.signing-name"]
				config.RestSigningRegion = catalog.Properties["rest.signing-region"]
			}
		}
	}

	// Set S3 flags - these are always true for S3 storage
	if config.S3Endpoint != "" {
		config.S3UseSSL = false
		config.S3PathStyle = true
	}

	return config, nil
}

func mapoptimisationTypeToOLake(optimisationType string) models.CatalogType {
	switch strings.ToLower(optimisationType) {
	case "custom":
		return "jdbc"
	case "glue":
		return "glue"
	case "rest":
		return "rest"
	case "hive":
		return "hive"
	default:
		return models.CatalogType(optimisationType)
	}
}

func normalizeCatalogType(olakeCatalogType string) string {
	switch strings.ToLower(olakeCatalogType) {
	case "glue", "rest", "hive":
		return olakeCatalogType
	case "jdbc":
		return "custom"
	default:
		return "custom"
	}
}

// setDefaultCatalogProperties sets required default properties for all catalogs
func setDefaultCatalogProperties(req *cmpModels.CatalogRequest) {
	if req.Properties == nil {
		req.Properties = make(map[string]string)
	}

	if _, exists := req.Properties["table.self-optimizing.enabled"]; !exists {
		req.Properties["table.self-optimizing.enabled"] = "false"
	}
	if _, exists := req.Properties["table.self-optimizing.quota"]; !exists {
		req.Properties["table.self-optimizing.quota"] = "0.1"
	}
	if _, exists := req.Properties["cache-enabled"]; !exists {
		req.Properties["cache-enabled"] = "false"
	}
	if _, exists := req.Properties["created-at"]; !exists {
		req.Properties["created-at"] = time.Now().Format("02 Jan 2006")
	}
}

func mapAuthConfig(olakeConfig *models.Config, authConfig, cmpStorageConfig map[string]string) {
	utils.SetIfNotEmpty(cmpStorageConfig, "storage.s3.region", olakeConfig.Region)
	utils.SetIfNotEmpty(cmpStorageConfig, "storage.s3.endpoint", olakeConfig.S3Endpoint)

	if olakeConfig.AccessKey != "" && olakeConfig.SecretKey != "" {
		authConfig["auth.type"] = "AK/SK"
		authConfig["auth.ak_sk.access_key"] = olakeConfig.AccessKey
		authConfig["auth.ak_sk.secret_key"] = olakeConfig.SecretKey
	} else {
		authConfig["auth.type"] = "CUSTOM"
	}
}

func mapCatalogProperties(olakeConfig *models.Config, properties map[string]string, olakeCatalogType string) {
	warehouse := olakeConfig.IcebergS3Path

	switch strings.ToLower(olakeCatalogType) {
	case "glue":
		properties["warehouse"] = warehouse

		if olakeConfig.UseGlueAdditionalConfig {
			utils.SetIfNotEmpty(properties, "glue.id", olakeConfig.GlueCatalogID)
			utils.SetIfNotEmpty(properties, "glue.endpoint", olakeConfig.GlueEndpoint)
			utils.SetIfNotEmpty(properties, "client.region", olakeConfig.GlueRegion)
		}
	case "jdbc":
		properties["warehouse"] = warehouse
		properties["catalog-impl"] = "org.apache.iceberg.jdbc.JdbcCatalog"
		utils.SetIfNotEmpty(properties, "uri", olakeConfig.JDBCUrl)
		utils.SetIfNotEmpty(properties, "jdbc.user", olakeConfig.JDBCUsername)
		utils.SetIfNotEmpty(properties, "jdbc.password", olakeConfig.JDBCPassword)
		utils.SetIfNotEmpty(properties, "endpoint", olakeConfig.S3Endpoint)
		utils.SetIfNotEmpty(properties, "access-key-id", olakeConfig.AccessKey)
		utils.SetIfNotEmpty(properties, "secret-access-key", olakeConfig.SecretKey)

	case "hive":
		utils.SetIfNotEmpty(properties, "warehouse", warehouse)
		utils.SetIfNotEmpty(properties, "uri", olakeConfig.HiveURI)
		if olakeConfig.HiveClients > 0 {
			utils.SetIfNotEmpty(properties, "clients", fmt.Sprintf("%d", olakeConfig.HiveClients))
		}

	case "rest":
		utils.SetIfNotEmpty(properties, "uri", olakeConfig.RestCatalogURL)
		utils.SetIfNotEmpty(properties, "warehouse", warehouse)
		utils.SetIfNotEmpty(properties, "token", olakeConfig.RestToken)
		utils.SetIfNotEmpty(properties, "oauth2-server-uri", olakeConfig.RestOAuthURI)
		utils.SetIfNotEmpty(properties, "rest.auth.type", olakeConfig.RestAuthType)
		utils.SetIfNotEmpty(properties, "credential", olakeConfig.RestCredential)
		utils.SetIfNotEmpty(properties, "scope", olakeConfig.RestScope)
		if olakeConfig.RestSigningV4 {
			utils.SetIfNotEmpty(properties, "rest.sigv4-enabled", "true")
			utils.SetIfNotEmpty(properties, "rest.signing-name", olakeConfig.RestSigningName)
			utils.SetIfNotEmpty(properties, "rest.signing-region", olakeConfig.RestSigningRegion)
		}
	}
}
