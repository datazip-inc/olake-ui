/**
 * Destination Connector Configuration Templates
 *
 * Reusable configuration templates for different destination connectors.
 * All destination fields are under the 'writer' prefix in RJSF.
 */

import { CatalogType, DestinationConnector } from "../enums"
import { DestinationFormConfig } from "../types/PageConfig.types"

/**
 * S3 Destination Configuration Template
 */
export const createS3DestinationConfig = (data: {
	name: string
	s3_bucket: string
	s3_region: string
	s3_path: string
	s3_endpoint?: string
	s3_access_key?: string
	s3_secret_key?: string
	version?: string
}): DestinationFormConfig => ({
	name: data.name,
	connector: DestinationConnector.AmazonS3,
	version: data.version,
	fields: {
		s3_bucket: data.s3_bucket,
		s3_region: data.s3_region,
		s3_path: data.s3_path,
		...(data.s3_endpoint && { s3_endpoint: data.s3_endpoint }),
		...(data.s3_access_key && { s3_access_key: data.s3_access_key }),
		...(data.s3_secret_key && { s3_secret_key: data.s3_secret_key }),
	},
})

/**
 * Iceberg with AWS Glue Catalog Configuration
 */
export const createIcebergGlueConfig = (data: {
	name: string
	aws_region: string
	iceberg_db: string
	iceberg_s3_path: string
	aws_access_key?: string
	aws_secret_key?: string
	s3_endpoint?: string
	version?: string
}): DestinationFormConfig => ({
	name: data.name,
	connector: DestinationConnector.ApacheIceberg,
	catalogType: CatalogType.Glue,
	version: data.version,
	fields: {
		aws_region: data.aws_region,
		iceberg_db: data.iceberg_db,
		iceberg_s3_path: data.iceberg_s3_path,
		...(data.aws_access_key && { aws_access_key: data.aws_access_key }),
		...(data.aws_secret_key && { aws_secret_key: data.aws_secret_key }),
		...(data.s3_endpoint && { s3_endpoint: data.s3_endpoint }),
	},
})

/**
 * Iceberg with JDBC Catalog Configuration
 */
export const createIcebergJdbcConfig = (data: {
	name: string
	jdbc_url: string
	jdbc_username: string
	jdbc_password: string
	iceberg_s3_path: string
	aws_region: string
	s3_endpoint?: string
	aws_access_key?: string
	aws_secret_key?: string
	s3_path_style?: boolean
	s3_use_ssl?: boolean
	version?: string
}): DestinationFormConfig => ({
	name: data.name,
	connector: DestinationConnector.ApacheIceberg,
	catalogType: CatalogType.JDBC,
	version: data.version,
	fields: {
		jdbc_url: data.jdbc_url,
		jdbc_username: data.jdbc_username,
		jdbc_password: data.jdbc_password,
		iceberg_s3_path: data.iceberg_s3_path,
		aws_region: data.aws_region,
		...(data.s3_endpoint && { s3_endpoint: data.s3_endpoint }),
		...(data.aws_access_key && { aws_access_key: data.aws_access_key }),
		...(data.aws_secret_key && { aws_secret_key: data.aws_secret_key }),
		...(data.s3_path_style !== undefined && {
			s3_path_style: data.s3_path_style,
		}),
		...(data.s3_use_ssl !== undefined && { s3_use_ssl: data.s3_use_ssl }),
	},
})

/**
 * Iceberg with Hive Catalog Configuration
 */
export const createIcebergHiveConfig = (data: {
	name: string
	hive_uri: string
	iceberg_db: string
	iceberg_s3_path: string
	aws_region: string
	hive_clients?: number
	hive_sasl_enabled?: boolean
	s3_endpoint?: string
	s3_path_style?: boolean
	s3_use_ssl?: boolean
	version?: string
}): DestinationFormConfig => ({
	name: data.name,
	connector: DestinationConnector.ApacheIceberg,
	catalogType: CatalogType.Hive,
	version: data.version,
	fields: {
		hive_uri: data.hive_uri,
		iceberg_db: data.iceberg_db,
		iceberg_s3_path: data.iceberg_s3_path,
		aws_region: data.aws_region,
		...(data.hive_clients && { hive_clients: data.hive_clients }),
		...(data.hive_sasl_enabled !== undefined && {
			hive_sasl_enabled: data.hive_sasl_enabled,
		}),
		...(data.s3_endpoint && { s3_endpoint: data.s3_endpoint }),
		...(data.s3_path_style !== undefined && {
			s3_path_style: data.s3_path_style,
		}),
		...(data.s3_use_ssl !== undefined && { s3_use_ssl: data.s3_use_ssl }),
	},
})

/**
 * Iceberg with REST Catalog Configuration
 */
export const createIcebergRestConfig = (data: {
	name: string
	rest_catalog_url: string
	iceberg_db: string
	iceberg_s3_path: string
	aws_region: string
	rest_auth_type?: string
	token?: string
	oauth2_uri?: string
	credential?: string
	scope?: string
	rest_signing_name?: string
	rest_signing_region?: string
	rest_signing_v_4?: boolean
	no_identifier_fields?: boolean
	s3_endpoint?: string
	version?: string
}): DestinationFormConfig => ({
	name: data.name,
	connector: DestinationConnector.ApacheIceberg,
	catalogType: CatalogType.Rest,
	version: data.version,
	fields: {
		rest_catalog_url: data.rest_catalog_url,
		iceberg_db: data.iceberg_db,
		iceberg_s3_path: data.iceberg_s3_path,
		aws_region: data.aws_region,
		...(data.rest_auth_type && { rest_auth_type: data.rest_auth_type }),
		...(data.token && { token: data.token }),
		...(data.oauth2_uri && { oauth2_uri: data.oauth2_uri }),
		...(data.credential && { credential: data.credential }),
		...(data.scope && { scope: data.scope }),
		...(data.rest_signing_name && {
			rest_signing_name: data.rest_signing_name,
		}),
		...(data.rest_signing_region && {
			rest_signing_region: data.rest_signing_region,
		}),
		...(data.rest_signing_v_4 !== undefined && {
			rest_signing_v_4: data.rest_signing_v_4,
		}),
		...(data.no_identifier_fields !== undefined && {
			no_identifier_fields: data.no_identifier_fields,
		}),
		...(data.s3_endpoint && { s3_endpoint: data.s3_endpoint }),
	},
})

/**
 * Test data configurations for E2E tests
 */

// S3 Destination Config
export const S3_DESTINATION_CONFIG = {
	bucketName: "test-bucket",
	region: "us-east-1",
	path: "s3://test-bucket/data",
	accessKey: "admin",
	secretKey: "password",
} as const

// Iceberg Destination Configs - All 4 Catalog Types

// 1. Iceberg with AWS Glue Catalog
export const ICEBERG_GLUE_CONFIG = {
	aws_region: "us-east-1",
	iceberg_s3_path: "s3a://warehouse",
	aws_access_key: "admin",
	aws_secret_key: "password",
	s3_endpoint: "http://172.17.0.2:9000",
} as const

// 2. Iceberg with JDBC Catalog
export const ICEBERG_JDBC_CONFIG = {
	jdbc_url: "jdbc:postgresql://172.17.0.2:5432/iceberg",
	jdbc_username: "iceberg",
	jdbc_password: "password",
	iceberg_s3_path: "s3a://warehouse",
	aws_region: "us-east-1",
	s3_endpoint: "http://172.17.0.2:9000",
	aws_access_key: "admin",
	aws_secret_key: "password",
	s3_use_ssl: false,
} as const
