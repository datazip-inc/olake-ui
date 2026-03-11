import { DESTINATION_INTERNAL_TYPES, DESTINATION_LABELS } from "../constants"
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
