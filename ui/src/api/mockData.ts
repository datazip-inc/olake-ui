import { Job, Source, Destination, StreamData, SourceJob } from "../types"

// Mock data for jobs
export const mockJobs: Job[] = [
	{
		id: "1",
		name: "Daily Sales Data Sync",
		status: "active",
		source: "MongoDB Sales DB",
		destination: "AWS S3 Data Lake",
		createdAt: new Date("2025-01-15T10:30:00Z"),
		lastSync: "2 hours ago",
		lastSyncStatus: "success",
		replication_frequency: "daily",
		paused: false,
	},
	{
		id: "2",
		name: "Inventory Sync",
		status: "active",
		source: "Postgres Inventory",
		destination: "AWS Glue Analytics",
		createdAt: new Date("2025-01-20T14:45:00Z"),
		lastSync: "5 hours ago",
		lastSyncStatus: "success",
		replication_frequency: "daily",
		paused: false,
	},
	{
		id: "3",
		name: "HR Data Backup",
		status: "inactive",
		source: "MySQL HR",
		destination: "REST Catalog Analytics",
		createdAt: new Date("2025-01-10T09:15:00Z"),
		lastSync: "1 day ago",
		lastSyncStatus: "failed",
		replication_frequency: "daily",
		paused: false,
	},
]

// Mock data for sources
export const mockSources: Source[] = [
	{
		id: "1",
		name: "MongoDB Sales DB",
		type: "MongoDB",
		status: "active",
		createdAt: new Date("2025-01-15T10:30:00Z"),
		config: {
			hosts: ["mongodb.example.com:27017"],
			username: "sales_admin",
			password: "sales_pass",
			authdb: "admin",
			"replica-set": "rs01",
			"read-preference": "secondaryPreferred",
			srv: false,
			"server-ram": 16,
			database: "sales_data",
			max_threads: 50,
			default_mode: "cdc",
			backoff_retry_count: 3,
			partition_strategy: "",
		},
		associatedJobs: [
			{
				source: "MongoDB Sales DB",
				destination: "AWS S3 Data Lake",
				jobName: "Daily Sales Data Sync",
			},
		],
	},
	{
		id: "2",
		name: "Postgres Inventory",
		type: "Postgres",
		status: "active",
		createdAt: new Date("2025-01-20T14:45:00Z"),
		config: {
			host: "postgres.example.com",
			port: 5432,
			database: "inventory_db",
			username: "inventory_user",
			password: "inventory_pass",
			jdbc_url_params: "connectTimeout=30",
			ssl: {
				mode: "disable",
			},
			update_method: {
				replication_slot: "inventory_slot",
				intial_wait_time: 10,
			},
			reader_batch_size: 100000,
			default_mode: "cdc",
			max_threads: 50,
		},
		associatedJobs: [
			{
				source: "Postgres Inventory",
				destination: "AWS Glue Analytics",
				jobName: "Inventory Sync",
			},
		],
	},
	{
		id: "3",
		name: "MySQL HR",
		type: "MySQL",
		status: "inactive",
		createdAt: new Date("2025-01-10T09:15:00Z"),
		config: {
			hosts: "mysql.example.com",
			port: 3306,
			database: "hr_system",
			username: "hr_admin",
			password: "hr_pass",
			update_method: {
				intial_wait_time: 10,
			},
			tls_skip_verify: true,
			max_threads: 10,
			backoff_retry_count: 2,
			default_mode: "cdc",
		},
		associatedJobs: [
			{
				source: "MySQL HR",
				destination: "REST Catalog Analytics",
				jobName: "HR Data Backup",
			},
		],
	},
]

// Mock data for destinations
export const mockDestinations: Destination[] = [
	{
		id: "1",
		name: "AWS S3 Data Lake",
		type: "Amazon S3",
		status: "active",
		createdAt: new Date("2025-01-15T10:30:00Z"),
		associatedJobs: [
			{
				source: "MongoDB Sales DB",
				destination: "AWS S3 Data Lake",
				jobName: "Daily Sales Data Sync",
			},
		],
	},
	{
		id: "2",
		name: "AWS Glue Analytics",
		type: "Apache Iceberg",
		catalog: "AWS Glue",
		status: "active",
		createdAt: new Date("2025-01-20T14:45:00Z"),
		associatedJobs: [
			{
				source: "Postgres Inventory",
				destination: "AWS Glue Analytics",
				jobName: "Inventory Sync",
			},
		],
	},
	{
		id: "3",
		name: "REST Catalog Analytics",
		type: "Apache Iceberg",
		catalog: "REST Catalog",
		status: "active",
		createdAt: new Date("2025-01-10T09:15:00Z"),
		associatedJobs: [
			{
				source: "MySQL HR",
				destination: "REST Catalog Analytics",
				jobName: "HR Data Backup",
			},
		],
	},
	{
		id: "4",
		name: "JDBC Catalog Analytics",
		type: "Apache Iceberg",
		catalog: "JDBC Catalog",
		status: "active",
		createdAt: new Date("2025-01-25T11:20:00Z"),
		associatedJobs: [
			{
				source: "MySQL HR",
				destination: "REST Catalog Analytics",
				jobName: "HR Data Backup",
			},
		],
	},
]

// Mock data for source jobs
export const mockSourceJobs: SourceJob[] = [
	{
		id: "1",
		name: "Daily Sales Data Sync",
		state: "running",
		lastRuntime: "2 hours ago",
		lastRuntimeStatus: "Success",
		destination: {
			name: "AWS S3 Data Lake",
			type: "AWS S3",
			config: {
				s3_bucket: "prod-data-lake",
				s3_region: "us-west-2",
				writer: "parquet",
			},
		},
		paused: false,
	},
	{
		id: "2",
		name: "Inventory Sync",
		state: "running",
		lastRuntime: "5 hours ago",
		lastRuntimeStatus: "Success",
		destination: {
			name: "AWS Glue Analytics",
			type: "AWS Glue Catalog",
			config: {
				database: "analytics_db",
				region: "us-west-2",
			},
		},
		paused: false,
	},
	{
		id: "3",
		name: "HR Data Backup",
		state: "stopped",
		lastRuntime: "1 day ago",
		lastRuntimeStatus: "Failed",
		destination: {
			name: "REST Catalog Analytics",
			type: "REST Catalog",
			config: {
				endpoint: "https://api.example.com/catalog",
				auth_type: "basic",
			},
		},
		paused: true,
	},
]

// Mock data for stream configurations
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
