package catalog

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/services/compaction/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	olake "github.com/datazip-inc/olake/destination/iceberg"
)

func MapOLakeConfigToCompactionCatalog(destinationName string, configJSON string) (*models.CatalogRequest, error) {
	var config olake.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse OLake config: %w", err)
	}

	catalogType := normalizeCatalogType(string(config.CatalogType))

	compactionReq := &models.CatalogRequest{
		Name:            destinationName,
		Type:            catalogType,
		OptimizerGroup:  models.OptimizerGroup,
		TableFormatList: "ICEBERG",
		StorageConfig:   make(map[string]string),
		AuthConfig:      make(map[string]string),
		Properties:      make(map[string]string),
		TableProperties: make(map[string]string),
	}

	compactionReq.StorageConfig["storage.type"] = models.DefaultStroageType

	mapAuthConfig(config, compactionReq.AuthConfig, compactionReq.StorageConfig)
	// Use the original OLake catalog type when deriving catalog-specific properties
	mapCatalogProperties(config, compactionReq.Properties, string(config.CatalogType))

	return compactionReq, nil
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

func mapAuthConfig(olakeConfig olake.Config, authConfig map[string]string, cmpStorageConfig map[string]string) {
	utils.SetIfNotEmpty(cmpStorageConfig, "storage.s3.region", olakeConfig.Region)
	utils.SetIfNotEmpty(cmpStorageConfig, "storage.s3.endpoint", olakeConfig.S3Endpoint)

	if olakeConfig.AccessKey != "" && olakeConfig.SecretKey != "" {
		// AK/SK is the auth type expected by Amoro
		authConfig["auth.type"] = "AK/SK"
		authConfig["auth.ak_sk.access_key"] = olakeConfig.AccessKey
		authConfig["auth.ak_sk.secret_key"] = olakeConfig.SecretKey
	} else {
		authConfig["auth.type"] = "CUSTOM"
	}
}

func mapCatalogProperties(olakeConfig olake.Config, properties map[string]string, olakeCatalogType string) {
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

		if strings.ToLower(string(olakeConfig.CatalogType)) == "jdbc" {
			utils.SetIfNotEmpty(properties, "uri", olakeConfig.JDBCUrl)
			utils.SetIfNotEmpty(properties, "jdbc.user", olakeConfig.JDBCUsername)
			utils.SetIfNotEmpty(properties, "jdbc.password", olakeConfig.JDBCPassword)
			properties["catalog-impl"] = "org.apache.iceberg.jdbc.JdbcCatalog"
		}

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
