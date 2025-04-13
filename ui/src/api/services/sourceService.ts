import api from "../axios"
import { Source } from "../../types"

// Flag to use mock data instead of real API
const useMockData = true

// Mock schemas for different connectors
const mockSourceConnectorSchemas = {
	MongoDB: {
		$schema: "https://json-schema.org/draft/2020-12/schema",
		type: "object",
		properties: {
			hosts: {
				type: "array",
				title: "Hosts",
				description: "List of MongoDB hosts. Use DNS SRV if srv = true.",
				items: {
					type: "string",
					title: "",
					pattern: "^[^\\s]+:\\d+$",
					errorMessage: {
						pattern: 'must match pattern "^[^\\s]+:\\d+$"',
					},
				},
				minItems: 1,
				uniqueItems: true,
				default: [""],
			},
			username: {
				type: "string",
				title: "Username",
				description: "Credentials for MongoDB authentication",
			},
			password: {
				type: "string",
				title: "Password",
				description: "Credentials for MongoDB authentication",
				format: "password",
			},
			authdb: {
				type: "string",
				title: "Auth Database",
				description: "Authentication database (often admin)",
				default: "admin",
			},
			"replica-set": {
				type: "string",
				title: "Replica Set",
				description: "Name of the replica set, if applicable",
			},
			"read-preference": {
				type: "string",
				title: "Read Preference",
				description: "Which node to read from (e.g., secondaryPreferred)",
				enum: [
					"primary",
					"primaryPreferred",
					"secondary",
					"secondaryPreferred",
					"nearest",
				],
				default: "secondaryPreferred",
			},
			srv: {
				type: "boolean",
				title: "SRV",
				description:
					"If using DNS SRV connection strings, set to true. When true, there can only be 1 host in hosts field.",
				default: false,
			},
			"server-ram": {
				type: "integer",
				title: "Server RAM",
				description: "Memory management hint for the OLake container",
				default: 16,
			},
			database: {
				type: "string",
				title: "Database",
				description: "The MongoDB database name to replicate",
			},
			max_threads: {
				type: "integer",
				title: "Max Threads",
				description: "Maximum parallel threads for chunk-based snapshotting",
				default: 50,
			},
			default_mode: {
				type: "string",
				title: "Default Mode",
				description: "Default sync mode",
				enum: ["cdc", "full_refresh", "incremental"],
				default: "cdc",
			},
			backoff_retry_count: {
				type: "integer",
				title: "Backoff Retry Count",
				description:
					"Retries attempt to establish sync again if it fails, increases exponentially (in minutes - 1, 2,4,8,16... depending upon the backoff_retry_count value)",
				default: 3,
			},
			partition_strategy: {
				type: "string",
				title: "Partition Strategy",
				description: "The partition strategy for backfill",
				enum: ["timestamp", ""],
				default: "",
			},
		},
		required: ["hosts", "database", "username", "password"],
		uiSchema: {
			"ui:order": [
				"hosts",
				"username",
				"password",
				"authdb",
				"replica-set",
				"read-preference",
				"srv",
				"server-ram",
				"database",
				"max_threads",
				"default_mode",
				"backoff_retry_count",
				"partition_strategy",
			],
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			hosts: {
				"ui:title": "Hosts",
				"ui:description": "List of MongoDB hosts. Use DNS SRV if srv = true.",
				"ui:field": "ArrayField",
				"ui:ArrayField": {
					"ui:addButtonText": "Add Host",
					"ui:orderable": false,
					"ui:removable": true,
					"ui:showTitle": false,
				},
				items: {
					"ui:emptyValue": "",
					"ui:placeholder": "Enter host:port (e.g. localhost:27017)",
					"ui:title": false,
					"ui:options": {
						label: false,
					},
				},
			},
		},
	},
	Postgres: {
		$schema: "https://json-schema.org/draft/2020-12/schema",
		type: "object",
		definitions: {
			SSLConfig: {
				properties: {
					mode: {
						enum: ["disable", "require", "verify-ca", "verify-full"],
						title: "SSL Mode",
						type: "string",
						description: "SSL mode for database connection",
						default: "disable",
					},
					client_cert: {
						title: "Client Certificate",
						type: "string",
						description: "Path to client certificate file",
					},
					client_key: {
						title: "Client Certificate Key",
						type: "string",
						description: "Path to client certificate key file",
					},
					server_ca: {
						title: "CA Certificate",
						type: "string",
						description: "Path to server CA certificate file",
					},
				},
				required: ["mode"],
				type: "object",
			},
			UpdateMethod: {
				type: "object",
				title: "Update Method Configuration",
				description: "Specifies the mechanism for updating data",
				properties: {
					replication_slot: {
						type: "string",
						title: "Replication Slot",
						description: "Name of the logical replication slot",
					},
					intial_wait_time: {
						type: "integer",
						title: "Initial Wait Time",
						description:
							"Time to wait before starting replication (in seconds)",
						default: 10,
					},
				},
				required: ["replication_slot", "intial_wait_time"],
			},
		},
		properties: {
			host: {
				type: "string",
				title: "Host",
				description: "The hostname or IP address of the database server",
				default: "localhost",
			},
			port: {
				type: "integer",
				title: "Port",
				description:
					"The port number through which the database server is accessible",
				default: 5432,
				minimum: 1,
				maximum: 65535,
			},
			database: {
				type: "string",
				title: "Database",
				description: "The name of the target database to connect to",
			},
			username: {
				type: "string",
				title: "Username",
				description: "The username used for authenticating with the database",
			},
			password: {
				type: "string",
				title: "Password",
				description:
					"The password corresponding to the provided username for authentication",
				format: "password",
			},
			jdbc_url_params: {
				type: "string",
				title: "JDBC URL Parameters",
				description:
					"A collection of additional JDBC URL parameters to fine-tune the connection",
			},
			ssl: {
				$ref: "#/definitions/SSLConfig",
				title: "SSL Configuration",
				description: "SSL configuration for the database connection",
			},
			update_method: {
				$ref: "#/definitions/UpdateMethod",
				title: "Update Method",
				description: "Specifies the mechanism for updating data",
			},
			reader_batch_size: {
				type: "integer",
				title: "Reader Batch Size",
				description:
					"The maximum number of records processed per batch during reading operations",
				default: 100000,
				minimum: 1,
			},
			default_mode: {
				type: "string",
				title: "Default Mode",
				description: "Defines the default mode of operation",
				enum: ["cdc", "full_refresh", "incremental"],
				default: "cdc",
			},
			max_threads: {
				type: "integer",
				title: "Max Threads",
				description:
					"The maximum number of threads allocated for parallel processing tasks",
				default: 50,
				minimum: 1,
			},
		},
		required: [
			"host",
			"port",
			"database",
			"username",
			"password",
			"ssl",
			"update_method",
		],
		uiSchema: {
			"ui:order": [
				"host",
				"port",
				"database",
				"username",
				"password",
				"ssl",
				"update_method",
				"reader_batch_size",
				"default_mode",
				"max_threads",
				"jdbc_url_params",
			],
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			ssl: {
				"ui:title": "SSL Configuration",
				"ui:description": "Configure SSL settings for the database connection",
				"ui:options": {
					className: "grid grid-cols-1 gap-4",
				},
			},
			update_method: {
				"ui:title": "Update Method Configuration",
				"ui:description": "Configure how data updates are handled",
				"ui:options": {
					className: "grid grid-cols-1 gap-4",
				},
			},
		},
	},
	MySQL: {
		$schema: "https://json-schema.org/draft/2020-12/schema",
		type: "object",
		definitions: {
			UpdateMethod: {
				type: "object",
				title: "Update Method Configuration",
				description: "Specifies the mechanism for updating data",
				properties: {
					intial_wait_time: {
						type: "integer",
						title: "Initial Wait Time",
						description:
							"Time to wait before starting replication (in seconds)",
						default: 10,
					},
				},
				required: ["intial_wait_time"],
			},
		},

		properties: {
			hosts: {
				type: "string",
				title: "Hosts",
				description: "List of database host addresses to connect to",
				default: "localhost",
			},
			username: {
				type: "string",
				title: "Username",
				description: "Username for authenticating with the database",
				default: "root",
			},
			password: {
				type: "string",
				title: "Password",
				description: "Password for the database user",
				format: "password",
			},
			database: {
				type: "string",
				title: "Database",
				description: "Name of the target database to use",
			},
			port: {
				type: "integer",
				title: "Port",
				description: "Port number on which the database server is listening",
				default: 3306,
				minimum: 1,
				maximum: 65535,
			},
			tls_skip_verify: {
				type: "boolean",
				title: "Skip TLS Verification",
				description:
					"Indicates whether to skip TLS certificate verification for secure connections",
				default: true,
			},
			update_method: {
				$ref: "#/definitions/UpdateMethod",
				title: "Update Method",
				description: "Specifies the mechanism for updating data",
			},
			default_mode: {
				type: "string",
				title: "Default Mode",
				description: "Default synchronization mode",
				enum: ["cdc", "full_refresh", "incremental"],
				default: "cdc",
			},
			max_threads: {
				type: "integer",
				title: "Max Threads",
				description:
					"Maximum number of parallel threads allowed for processing or syncing data",
				default: 10,
				minimum: 1,
			},
			backoff_retry_count: {
				type: "integer",
				title: "Backoff Retry Count",
				description:
					"Number of retry attempts for establishing sync, using an exponential backoff strategy upon failures",
				default: 2,
				minimum: 0,
			},
		},
		required: [
			"hosts",
			"username",
			"password",
			"database",
			"port",
			"update_method",
		],
		uiSchema: {
			"ui:order": [
				"hosts",
				"port",
				"database",
				"username",
				"password",
				"tls_skip_verify",
				"update_method",
				"default_mode",
				"max_threads",
				"backoff_retry_count",
			],
			"ui:className":
				"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
			password: {
				"ui:widget": "password",
			},
			tls_skip_verify: {
				"ui:widget": "checkbox",
				"ui:options": {
					inline: true,
				},
			},
			update_method: {
				// "ui:title": "Update Method Configuration",
				"ui:description": "Configure how data updates are handled",
				"ui:options": {
					className: "grid grid-cols-1 gap-4",
				},
			},
		},
	},
}

// Mock data for sources
const mockSources: Source[] = [
	{
		id: "1",
		name: "MongoDB Source",
		type: "MongoDB",
		status: "active",
		createdAt: new Date("2025-01-15T10:30:00Z"),
		config: {
			hosts: ["localhost:27017"],
			username: "admin",
			password: "password",
			authdb: "admin",
			"replica-set": "rs01",
			"read-preference": "secondaryPreferred",
			srv: false,
			"server-ram": 16,
			database: "test_db",
			collection: "test_collection",
			max_threads: 50,
			default_mode: "cdc",
			backoff_retry_count: 3,
			partition_strategy: "",
		},
	},
	{
		id: "2",
		name: "PostgreSQL Source",
		type: "PostgreSQL",
		status: "active",
		createdAt: new Date("2025-01-20T14:45:00Z"),
		config: {
			host: "localhost",
			port: 5432,
			database: "test_db",
			username: "postgres",
			password: "password",
			jdbc_url_params: "connectTimeout=30",
			schema: "public",
			table: "test_table",
			ssl: {
				mode: "disable",
				client_cert: "",
				client_key: "",
				server_ca: "",
			},
			update_method: {
				replication_slot: "test_slot",
				intial_wait_time: 10,
			},
			publication: "test_publication",
		},
	},
	{
		id: "3",
		name: "MySQL Source",
		type: "MySQL",
		status: "active",
		createdAt: new Date("2025-01-10T09:15:00Z"),
		config: {
			hosts: "localhost",
			port: 3306,
			database: "test_db",
			username: "root",
			password: "password",
			table: "test_table",
			tls_skip_verify: true,
			update_method: {
				intial_wait_time: 10,
			},
			server_id: 1,
			binlog_position: "mysql-bin.000001:1234",
			max_threads: 10,
			backoff_retry_count: 2,
			default_mode: "cdc",
		},
	},
]

export const sourceService = {
	// Get all sources
	getSources: async () => {
		if (useMockData) {
			// Return mock data with a small delay to simulate network request
			return new Promise<Source[]>(resolve => {
				setTimeout(() => resolve(mockSources), 500)
			})
		}

		const response = await api.get<Source[]>("/sources")
		return response.data
	},

	// Get source by id
	getSourceById: async (id: string) => {
		if (useMockData) {
			const source = mockSources.find(source => source.id === id)
			if (!source) throw new Error("Source not found")

			return new Promise<Source>(resolve => {
				setTimeout(() => resolve(source), 300)
			})
		}

		const response = await api.get<Source>(`/sources/${id}`)
		return response.data
	},

	// Create new source
	createSource: async (source: Omit<Source, "id" | "createdAt">) => {
		if (useMockData) {
			const newSource: Source = {
				...source,
				id: Math.random().toString(36).substring(2, 9),
				createdAt: new Date(),
			}

			// Only add to the local mockSources array to avoid duplication
			mockSources.push(newSource)

			return new Promise<Source>(resolve => {
				setTimeout(() => resolve(newSource), 400)
			})
		}

		const response = await api.post<Source>("/sources", source)
		return response.data
	},

	// Update source
	updateSource: async (id: string, source: Partial<Source>) => {
		if (useMockData) {
			const index = mockSources.findIndex(s => s.id === id)
			if (index === -1) throw new Error("Source not found")

			const updatedSource = { ...mockSources[index], ...source }
			mockSources[index] = updatedSource

			return new Promise<Source>(resolve => {
				setTimeout(() => resolve(updatedSource), 300)
			})
		}

		const response = await api.put<Source>(`/sources/${id}`, source)
		return response.data
	},

	// Delete source
	deleteSource: async (id: string) => {
		if (useMockData) {
			const index = mockSources.findIndex(s => s.id === id)
			if (index === -1) throw new Error("Source not found")

			mockSources.splice(index, 1)

			return new Promise<void>(resolve => {
				setTimeout(() => resolve(), 300)
			})
		}

		const response = await api.delete(`/sources/${id}`)
		return response.data
	},

	// Test source connection
	testConnection: async (id: string) => {
		if (useMockData) {
			const source = mockSources.find(s => s.id === id)
			if (!source) throw new Error("Source not found")

			return new Promise<{ success: boolean; message: string }>(resolve => {
				setTimeout(
					() =>
						resolve({
							success: true,
							message: "Connection successful",
						}),
					800,
				)
			})
		}

		const response = await api.post(`/sources/${id}/test`)
		return response.data
	},

	getConnectorSchema: async (connectorType: string) => {
		try {
			if (useMockData) {
				// Return mock schema for development
				const schema =
					mockSourceConnectorSchemas[
						connectorType as keyof typeof mockSourceConnectorSchemas
					]
				// If schema doesn't exist, return empty object
				if (!schema) {
					return {}
				}
				return { ...schema }
			}

			// For production: fetch schema from API
			const response = await api.get(`/connectors/${connectorType}/schema`)
			return response.data
		} catch (error) {
			console.error(`Error fetching schema for ${connectorType}:`, error)
			throw error
		}
	},
}
