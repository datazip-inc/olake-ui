import {
	GitCommitIcon,
	LinktreeLogoIcon,
	PathIcon,
} from "@phosphor-icons/react"
import { NavItem, TestConnectionStatus } from "../types"

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

export const THEME_CONFIG = {
	token: {
		colorPrimary: "#203FDD",
		borderRadius: 6,
	},
}

export const NAV_ITEMS: NavItem[] = [
	{ path: "/jobs", label: "Jobs", icon: GitCommitIcon },
	{ path: "/sources", label: "Sources", icon: LinktreeLogoIcon },
	{ path: "/destinations", label: "Destinations", icon: PathIcon },
]

// not showing oneof and const errors
export const transformErrors = (errors: any[]) => {
	return errors.filter(err => err.name !== "oneOf" && err.name !== "const")
}

export const TEST_CONNECTION_STATUS: Record<TestConnectionStatus, string> = {
	SUCCEEDED: "SUCCEEDED",
	FAILED: "FAILED",
} as const

export const SourceTutorialYTLink =
	"https://www.youtube.com/watch?v=dQw4w9WgXcQ"
export const DestinationTutorialYTLink =
	"https://www.youtube.com/watch?v=dQw4w9WgXcQ"
