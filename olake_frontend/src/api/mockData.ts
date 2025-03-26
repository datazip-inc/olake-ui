import { Job, Source, Destination, StreamData } from "../types"

// Mock data for jobs
export const mockJobs: Job[] = [
	{
		id: "1",
		name: "Daily Sales Data Sync",
		status: "active",
		source: "MongoDB Sales DB",
		destination: "Snowflake Data Warehouse",
		createdAt: new Date("2025-01-15T10:30:00Z"),
		lastSync: "2 hours ago",
		lastSyncStatus: "success",
	},
	{
		id: "2",
		name: "User Analytics Pipeline",
		status: "active",
		source: "Kafka User Events",
		destination: "BigQuery Analytics",
		createdAt: new Date("2025-01-20T14:45:00Z"),
		lastSync: "5 hours ago",
		lastSyncStatus: "success",
	},
	{
		id: "3",
		name: "Inventory Sync",
		status: "inactive",
		source: "PostgreSQL Inventory",
		destination: "Amazon S3 Data Lake",
		createdAt: new Date("2025-01-10T09:15:00Z"),
		lastSync: "1 day ago",
		lastSyncStatus: "failed",
	},
	{
		id: "4",
		name: "Marketing Campaign Data",
		status: "saved",
		source: "REST API Marketing",
		destination: "Redshift Reporting",
		createdAt: new Date("2025-01-25T11:20:00Z"),
		lastSync: "Never",
		lastSyncStatus: undefined,
	},
	{
		id: "5",
		name: "Customer Support Tickets",
		status: "failed",
		source: "MongoDB Support",
		destination: "Snowflake Data Warehouse",
		createdAt: new Date("2025-01-22T16:10:00Z"),
		lastSync: "3 hours ago",
		lastSyncStatus: "failed",
	},
	{
		id: "6",
		name: "Product Catalog Sync",
		status: "active",
		source: "MySQL Products",
		destination: "Amazon S3 Catalog",
		createdAt: new Date("2025-01-18T13:25:00Z"),
		lastSync: "6 hours ago",
		lastSyncStatus: "success",
	},
	{
		id: "7",
		name: "Financial Transactions ETL",
		status: "active",
		source: "PostgreSQL Finance",
		destination: "BigQuery Finance",
		createdAt: new Date("2025-01-14T08:40:00Z"),
		lastSync: "12 hours ago",
		lastSyncStatus: "success",
	},
	{
		id: "8",
		name: "Website Logs Analysis",
		status: "active",
		source: "Kafka Web Logs",
		destination: "Snowflake Analytics",
		createdAt: new Date("2025-01-19T15:50:00Z"),
		lastSync: "1 hour ago",
		lastSyncStatus: "running",
	},
	{
		id: "9",
		name: "HR Data Backup",
		status: "inactive",
		source: "MySQL HR",
		destination: "Amazon S3 Backup",
		createdAt: new Date("2025-01-12T10:15:00Z"),
		lastSync: "2 days ago",
		lastSyncStatus: "success",
	},
	{
		id: "10",
		name: "IoT Sensor Data Collection",
		status: "active",
		source: "MongoDB IoT",
		destination: "Redshift IoT Analytics",
		createdAt: new Date("2025-01-21T09:30:00Z"),
		lastSync: "30 minutes ago",
		lastSyncStatus: "success",
	},
]

// Mock data for sources
export const mockSources: Source[] = [
	// MongoDB Sources
	{
		id: "mongo-1",
		name: "MongoDB Sales DB",
		type: "MongoDB",
		status: "active",
		createdAt: new Date("2025-01-05T08:30:00Z"),
	},
	{
		id: "mongo-2",
		name: "MongoDB Support",
		type: "MongoDB",
		status: "active",
		createdAt: new Date("2025-01-07T14:10:00Z"),
	},
	{
		id: "mongo-3",
		name: "MongoDB IoT",
		type: "MongoDB",
		status: "active",
		createdAt: new Date("2025-01-11T11:45:00Z"),
	},
	{
		id: "mongo-4",
		name: "MongoDB Customer Data",
		type: "MongoDB",
		status: "inactive",
		createdAt: new Date("2025-01-09T16:20:00Z"),
	},

	// PostgreSQL Sources
	{
		id: "postgres-1",
		name: "PostgreSQL Inventory",
		type: "PostgreSQL",
		status: "inactive",
		createdAt: new Date("2025-01-03T10:15:00Z"),
	},
	{
		id: "postgres-2",
		name: "PostgreSQL Finance",
		type: "PostgreSQL",
		status: "active",
		createdAt: new Date("2025-01-06T09:40:00Z"),
	},
	{
		id: "postgres-3",
		name: "PostgreSQL Orders",
		type: "PostgreSQL",
		status: "saved",
		createdAt: new Date("2025-01-08T13:25:00Z"),
	},

	// MySQL Sources
	{
		id: "mysql-1",
		name: "MySQL Products",
		type: "MySQL",
		status: "active",
		createdAt: new Date("2025-01-04T11:35:00Z"),
	},
	{
		id: "mysql-2",
		name: "MySQL HR",
		type: "MySQL",
		status: "inactive",
		createdAt: new Date("2025-01-02T15:50:00Z"),
	},

	// Kafka Sources
	{
		id: "kafka-1",
		name: "Kafka User Events",
		type: "Kafka",
		status: "active",
		createdAt: new Date("2025-01-08T13:45:00Z"),
	},
	{
		id: "kafka-2",
		name: "Kafka Web Logs",
		type: "Kafka",
		status: "active",
		createdAt: new Date("2025-01-10T10:30:00Z"),
	},

	// REST API Sources
	{
		id: "rest-1",
		name: "REST API Marketing",
		type: "REST API",
		status: "saved",
		createdAt: new Date("2025-01-12T09:20:00Z"),
	},
]

// Mock data for destinations
export const mockDestinations: Destination[] = [
	// Snowflake Destinations
	{
		id: "snowflake-1",
		name: "Snowflake Data Warehouse",
		type: "Snowflake",
		status: "active",
		createdAt: new Date("2025-01-02T11:30:00Z"),
	},
	{
		id: "snowflake-2",
		name: "Snowflake Analytics",
		type: "Snowflake",
		status: "active",
		createdAt: new Date("2025-01-05T14:15:00Z"),
	},
	{
		id: "snowflake-3",
		name: "Snowflake Reporting",
		type: "Snowflake",
		status: "inactive",
		createdAt: new Date("2025-01-07T09:45:00Z"),
	},

	// BigQuery Destinations
	{
		id: "bigquery-1",
		name: "BigQuery Analytics",
		type: "BigQuery",
		status: "active",
		createdAt: new Date("2025-01-04T15:45:00Z"),
	},
	{
		id: "bigquery-2",
		name: "BigQuery Finance",
		type: "BigQuery",
		status: "active",
		createdAt: new Date("2025-01-06T10:20:00Z"),
	},
	{
		id: "bigquery-3",
		name: "BigQuery Marketing",
		type: "BigQuery",
		status: "saved",
		createdAt: new Date("2025-01-08T13:10:00Z"),
	},

	// Amazon S3 Destinations
	{
		id: "s3-1",
		name: "Amazon S3 Data Lake",
		type: "Amazon S3",
		status: "inactive",
		createdAt: new Date("2025-01-01T09:15:00Z"),
	},
	{
		id: "s3-2",
		name: "Amazon S3 Catalog",
		type: "Amazon S3",
		status: "active",
		createdAt: new Date("2025-01-03T14:30:00Z"),
	},
	{
		id: "s3-3",
		name: "Amazon S3 Backup",
		type: "Amazon S3",
		status: "active",
		createdAt: new Date("2025-01-05T11:25:00Z"),
	},

	// Redshift Destinations
	{
		id: "redshift-1",
		name: "Redshift Reporting",
		type: "Redshift",
		status: "saved",
		createdAt: new Date("2025-01-06T12:20:00Z"),
	},
	{
		id: "redshift-2",
		name: "Redshift IoT Analytics",
		type: "Redshift",
		status: "active",
		createdAt: new Date("2025-01-09T10:40:00Z"),
	},
]

export const mockStreamData: StreamData[] = [
	{
		sync_mode: "full_refresh",
		destination_sync_mode: "overwrite",
		selected_columns: null,
		sort_key: ["eventn_ctx_event_id"],
		stream: {
			json_schema: {
				properties: {
					"canonical-vid": {
						type: ["null", "integer"],
					},
					"internal-list-id": {
						type: ["null", "integer"],
					},
					"is-member": {
						type: ["null", "boolean"],
					},
					"static-list-id": {
						type: ["null", "integer"],
					},
					timestamp: {
						type: ["null", "integer"],
					},
					vid: {
						type: ["null", "integer"],
					},
				},
			},
			name: "contacts_list_memberships",
			source_defined_cursor: false,
			supported_sync_modes: ["full_refresh"],
		},
	},
	{
		sync_mode: "cdc",
		cursor_field: ["updatedAt"],
		destination_sync_mode: "overwrite",
		selected_columns: null,
		sort_key: null,
		stream: {
			default_cursor_field: ["updatedAt"],
			json_schema: {
				properties: {
					archived: {
						type: ["null", "boolean"],
					},
					companies: {
						type: ["null", "array"],
					},
					contacts: {
						type: ["null", "array"],
					},
					createdAt: {
						format: "date-time",
						type: ["null", "string"],
					},
					id: {
						type: ["null", "string"],
					},
					line_items: {
						type: ["null", "array"],
					},
					properties: {
						properties: {
							amount: {
								type: ["null", "number"],
							},
						},
						type: "object",
					},
					updatedAt: {
						format: "date-time",
						type: ["null", "string"],
					},
				},
			},
			name: "deals",
			source_defined_cursor: true,
			source_defined_primary_key: [["id"]],
			supported_sync_modes: ["full_refresh", "incremental"],
		},
	},
]
