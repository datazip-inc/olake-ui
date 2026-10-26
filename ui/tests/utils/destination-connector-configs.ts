/**
 * Destination Connector Configuration Templates
 *
 * Reusable configuration templates for different destination connectors.
 * All destination fields are under the 'writer' prefix in RJSF.
 */

import { CatalogType, DestinationConnector } from "../enums"
import { SERVICE_HOST } from "./constants"
import { DestinationFormConfig } from "../types/PageConfig.types"

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
 * The Iceberg destination of the integration test: JDBC catalog in iceberg-postgres, data in
 * MinIO (server/tests/destinations/jdbc/docker-compose.yml)
 */
export const ICEBERG_JDBC_CONFIG = {
	jdbc_url: `jdbc:postgresql://${SERVICE_HOST}:5432/iceberg`,
	jdbc_username: "iceberg",
	jdbc_password: "password",
	iceberg_s3_path: "s3a://warehouse",
	aws_region: "us-east-1",
	s3_endpoint: `http://${SERVICE_HOST}:9000`,
	aws_access_key: "admin",
	aws_secret_key: "password",
	s3_use_ssl: false,
} as const
