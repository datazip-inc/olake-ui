import { AxiosError } from "axios"

import {
	AWSS3,
	ApacheIceBerg,
	DB2,
	Kafka,
	MongoDB,
	MySQL,
	Oracle,
	Postgres,
	MSSQL,
} from "@/assets"
import { OperationError } from "@/common/services/operationsService"

import { DESTINATION_INTERNAL_TYPES, DESTINATION_LABELS } from "../constants"

// Normalizes old connector types to their current internal types
export const normalizeConnectorType = (connectorType: string): string => {
	const lowerType = connectorType.toLowerCase()

	switch (lowerType) {
		case "s3":
			return "s3"
		case "amazon s3":
			return "parquet"
		case "iceberg":
		case "apache iceberg":
			return "iceberg"
		default:
			return connectorType
	}
}

// These are used to show in connector dropdowns
export const getConnectorImage = (connector: string) => {
	const normalizedConnector = normalizeConnectorType(connector).toLowerCase()

	switch (normalizedConnector) {
		case "mongodb":
			return MongoDB
		case "postgres":
			return Postgres
		case "mysql":
			return MySQL
		case "oracle":
			return Oracle
		case DESTINATION_INTERNAL_TYPES.S3:
			return AWSS3
		case DESTINATION_INTERNAL_TYPES.ICEBERG:
			return ApacheIceBerg
		case "kafka":
			return Kafka
		case "s3":
			return AWSS3
		case "db2":
			return DB2
		case "mssql":
			return MSSQL
		default:
			// Default placeholder
			return MongoDB
	}
}

export const getConnectorInLowerCase = (connector?: string | null) => {
	const normalizedConnector = normalizeConnectorType(connector || "")
	const lowerConnector = normalizedConnector.toLowerCase()

	switch (lowerConnector) {
		case DESTINATION_INTERNAL_TYPES.S3:
		case DESTINATION_LABELS.AMAZON_S3:
			return DESTINATION_INTERNAL_TYPES.S3
		case DESTINATION_INTERNAL_TYPES.ICEBERG:
		case DESTINATION_LABELS.APACHE_ICEBERG:
			return DESTINATION_INTERNAL_TYPES.ICEBERG
		case "s3":
			return "s3"
		case "mongodb":
			return "mongodb"
		case "postgres":
			return "postgres"
		case "mysql":
			return "mysql"
		case "oracle":
			return "oracle"
		case "db2":
			return "db2"
		case "mssql":
			return "mssql"
		default:
			return lowerConnector
	}
}

/**
 * Extracts a message to show when a connection test could not produce a verdict.
 *
 * A connector that runs and reports it cannot reach the database is a *successful*
 * operation carrying a FAILED status, and never reaches here. This covers the other case:
 * the workflow itself did not complete — a failed image pull, a timeout, a lost
 * connection while polling. Those surface as OperationError, so unwrapping only
 * AxiosError would reduce a real, actionable message to "Unknown error occurred".
 */
export const describeConnectionFailure = (error: unknown): string => {
	if (error instanceof OperationError) {
		return error.message || "Connection test failed"
	}
	if (error instanceof AxiosError) {
		return (
			error.response?.data?.message ??
			"Network error - please check your connection"
		)
	}
	if (error instanceof Error && error.message) {
		return error.message
	}
	return "Unknown error occurred"
}
