export const CONNECTOR_TYPES = {
	AMAZON_S3: "Amazon S3",
	APACHE_ICEBERG: "Apache Iceberg",
	MONGODB: "MongoDB",
	POSTGRES: "Postgres",
	MYSQL: "MySQL",
	ORACLE: "Oracle",
	KAFKA: "Kafka",
	S3: "Amazon S3",
	DB2: "DB2",
	MSSQL: "MSSQL",
	DESTINATION_DEFAULT_CONNECTOR: "Amazon S3",
	SOURCE_DEFAULT_CONNECTOR: "MongoDB",
}

export const SETUP_TYPES = {
	NEW: "new",
	EXISTING: "existing",
} as const

export const STATUS = {
	ACTIVE: "active",
	INACTIVE: "inactive",
	PENDING: "pending",
	FAILED: "failed",
	RUNNING: "running",
	SUCCESS: "success",
	COMPLETED: "completed",
	CANCELLED: "cancelled",
}

export const STATUS_LABELS = {
	ACTIVE: "Active",
	INACTIVE: "Inactive",
	PENDING: "Pending",
	FAILED: "Failed",
	RUNNING: "Running",
	SUCCESS: "Success",
	COMPLETED: "Completed",
	CANCELLED: "Cancelled",
}
export const ENTITY_TYPES = {
	SOURCE: "source",
	DESTINATION: "destination",
	JOB: "job",
}

export const DESTINATION_INTERNAL_TYPES = {
	ICEBERG: "iceberg",
	S3: "parquet",
}

export const SOURCE_INTERNAL_TYPES = {
	MONGODB: "mongodb",
	POSTGRES: "postgres",
	MYSQL: "mysql",
	ORACLE: "oracle",
	KAFKA: "kafka",
	S3: "s3",
	DB2: "db2",
	MSSQL: "mssql",
} as const

export const DESTINATION_LABELS = {
	AMAZON_S3: "amazon s3",
	APACHE_ICEBERG: "apache iceberg",
}

export const PAGE_SIZE = window.innerHeight >= 900 ? 8 : 6

export const DISPLAYED_JOBS_COUNT = 5

// Minimum source version that supports column selection.
export const MIN_COLUMN_SELECTION_SOURCE_VERSION = "v0.4.0"
