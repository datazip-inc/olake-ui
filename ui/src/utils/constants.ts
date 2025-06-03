import { GitCommit, LinktreeLogo, Path } from "@phosphor-icons/react"
import { CatalogType, NavItem } from "../types"

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
}

export const CATALOG_TYPES = {
	AWS_GLUE: "AWS Glue",
	REST_CATALOG: "REST Catalog",
	JDBC_CATALOG: "JDBC Catalog",
	HIVE_CATALOG: "Hive Catalog",
	NONE: "None",
}

export const SETUP_TYPES = {
	NEW: "new",
	EXISTING: "existing",
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
	{ key: "active", label: "Active sources" },
	{ key: "inactive", label: "Inactive sources" },
]

export const mapCatalogValueToType = (
	catalogValue: string,
): CatalogType | null => {
	if (catalogValue === "none") return CATALOG_TYPES.NONE
	if (catalogValue === "glue") return CATALOG_TYPES.AWS_GLUE
	if (catalogValue === "rest") return CATALOG_TYPES.REST_CATALOG
	if (catalogValue === "jdbc") return CATALOG_TYPES.JDBC_CATALOG
	if (catalogValue === "hive") return CATALOG_TYPES.HIVE_CATALOG
	return null
}

export const IcebergCatalogTypes = [
	{ value: CATALOG_TYPES.AWS_GLUE, label: "AWS Glue" },
	{ value: CATALOG_TYPES.REST_CATALOG, label: "REST catalog" },
	{ value: CATALOG_TYPES.JDBC_CATALOG, label: "JDBC Catalog" },
	{ value: CATALOG_TYPES.HIVE_CATALOG, label: "Hive Catalog" },
]

export const destinationTabs = [
	{ key: "active", label: "Active destinations" },
	{ key: "inactive", label: "Inactive destinations" },
]
