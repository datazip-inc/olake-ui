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
