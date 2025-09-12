import { GitCommit, LinktreeLogo, Path } from "@phosphor-icons/react"
import { JobCreationSteps, NavItem } from "../types"

export const PARTITIONING_COLUMNS = [
	{
		title: "Column name",
		dataIndex: "name",
		key: "name",
	},
	{
		title: "Granularity",
		dataIndex: "granularity",
		key: "granularity",
	},
	{
		title: "Default",
		dataIndex: "default",
		key: "default",
	},
]

export const CONNECTOR_TYPES = {
	AMAZON_S3: "Amazon S3",
	APACHE_ICEBERG: "Apache Iceberg",
	MONGODB: "MongoDB",
	POSTGRES: "Postgres",
	MYSQL: "MySQL",
	ORACLE: "Oracle",
	DESTINATION_DEFAULT_CONNECTOR: "Amazon S3",
	SOURCE_DEFAULT_CONNECTOR: "MongoDB",
}

export const SETUP_TYPES = {
	NEW: "new",
	EXISTING: "existing",
}

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

export const JOB_CREATION_STEPS: Record<string, JobCreationSteps> = {
	SOURCE: "source",
	DESTINATION: "destination",
	STREAMS: "streams",
	CONFIG: "config",
} as const

export const TAB_TYPES = {
	CONFIG: "config",
	SCHEMA: "schema",
	JOBS: "jobs",
}

export const ENTITY_TYPES = {
	SOURCE: "source",
	DESTINATION: "destination",
	JOB: "job",
}

export const DESTINATION_INTERNAL_TYPES = {
	ICEBERG: "iceberg",
	S3: "s3",
}

export const DESTINATION_LABELS = {
	AMAZON_S3: "amazon s3",
	APACHE_ICEBERG: "apache iceberg",
}

export const JOB_TYPES = {
	ACTIVE: "active",
	INACTIVE: "inactive",
	SAVED: "saved",
	FAILED: "failed",
}

export const PAGE_SIZE = 8

export const THEME_CONFIG = {
	token: {
		colorPrimary: "#203FDD",
		borderRadius: 6,
	},
}

export const HTTP_STATUS = {
	UNAUTHORIZED: 401,
	FORBIDDEN: 403,
	SERVER_ERROR: 500,
}

export const ERROR_MESSAGES = {
	AUTH_REQUIRED: "Authentication required. Please log in.",
	NO_PERMISSION: "You do not have permission to access this resource",
	SERVER_ERROR: "Server error occurred. Please try again later.",
	NO_RESPONSE:
		"No response received from server. Please check your connection.",
}

export const LOCALSTORAGE_TOKEN_KEY = "token"
export const LOCALSTORAGE_USERNAME_KEY = "username"

export const NAV_ITEMS: NavItem[] = [
	{ path: "/jobs", label: "Jobs", icon: GitCommit },
	{ path: "/sources", label: "Sources", icon: LinktreeLogo },
	{ path: "/destinations", label: "Destinations", icon: Path },
]

export const sourceTabs = [
	{ key: STATUS.ACTIVE, label: "Active sources" },
	{ key: STATUS.INACTIVE, label: "Inactive sources" },
]

export const destinationTabs = [
	{ key: STATUS.ACTIVE, label: "Active destinations" },
	{ key: STATUS.INACTIVE, label: "Inactive destinations" },
]

export const COLORS = {
	selected: {
		border: "#203FDD",
		text: "#203FDD",
	},
	unselected: {
		border: "#D9D9D9",
		text: "#575757",
	},
} as const

export const steps: string[] = [
	JOB_CREATION_STEPS.CONFIG,
	JOB_CREATION_STEPS.SOURCE,
	JOB_CREATION_STEPS.DESTINATION,
	JOB_CREATION_STEPS.STREAMS,
]

export const TAB_STYLES = {
	active: "border border-primary bg-white text-primary rounded-md py-1 px-2",
	inactive: "bg-background-primary text-slate-900 py-1 px-2",
	hover: "hover:text-primary",
}

export const CARD_STYLE = "rounded-xl border border-[#E3E3E3] p-3"

export const JobTutorialYTLink =
	"https://youtu.be/_qRulFv-BVM?si=NPTw9V0hWQ3-9wOP"
export const SourceTutorialYTLink =
	"https://youtu.be/ndCHGlK5NCM?si=jvPy-aMrpEXCQA-8"
export const DestinationTutorialYTLink =
	"https://youtu.be/Ub1pcLg0WsM?si=V2tEtXvx54wDoa8Y"

export const connectorTypeMap: Record<string, string> = {
	mongodb: "MongoDB",
	postgres: "Postgres",
	mysql: "MySQL",
	oracle: "Oracle",
}

export const DAYS_MAP = {
	Sunday: 0,
	Monday: 1,
	Tuesday: 2,
	Wednesday: 3,
	Thursday: 4,
	Friday: 5,
	Saturday: 6,
}

export const DAYS = [
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
]

export const FREQUENCY_OPTIONS = [
	{ value: "minutes", label: "Every Minute" },
	{ value: "hours", label: "Every Hour" },
	{ value: "days", label: "Every Day" },
	{ value: "weeks", label: "Every Week" },
	{ value: "custom", label: "Custom" },
]

export const PartitioningRegexTooltip =
	"Choose a column to partition your data for faster reads and better performance"

export const DISPLAYED_JOBS_COUNT = 5

export const SYNC_MODE_MAP = {
	FULL_REFRESH: "full_refresh",
	INCREMENTAL: "incremental",
	CDC: "cdc",
	STRICT_CDC: "strict_cdc",
}

export const JOB_STEP_NUMBERS = {
	CONFIG: 1,
	SOURCE: 2,
	DESTINATION: 3,
	STREAMS: 4,
} as const

export const transformErrors = (errors: any[]) => {
	return errors.filter(err => err.name !== "oneOf" && err.name !== "const")
}

export const FORMAT_OPTIONS = {
	DYNAMIC: "dynamic",
	CUSTOM: "custom",
} as const

export const NAMESPACE_PLACEHOLDER = "_${source_namespace}"

export const LABELS = {
	S3: {
		title: "S3 Folder Name",
		folderType: "S3",
	},
	ICEBERG: {
		title: "Iceberg Database Name",
		folderType: "Iceberg DB",
	},
} as const
