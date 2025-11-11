/**
 * Source Connector Configuration Templates
 *
 * This file contains reusable configuration templates for different source connectors.
 * When adding a new connector, simply add a new configuration here - no need to modify
 * the CreateSourcePage class.
 *
 * @module SourceConnectorConfigs
 */

import { SourceConnector } from "../enums"
import { SourceFormConfig } from "../types/PageConfig.types"

/**
 * PostgreSQL Source Configuration Template
 * Based on the actual Postgres connector schema
 */
export const createPostgresSourceConfig = (data: {
	name: string
	host: string
	port: string | number
	database: string
	username: string
	password: string
	version?: string
	// Optional advanced fields
	ssl_mode?: "require" | "disable" | "verify-ca" | "verify-full"
	max_threads?: number
	retry_count?: number
	jdbc_url_params?: Record<string, any>
	update_method?: "Standalone" | "CDC"
	// CDC-specific fields
	replication_slot?: string
	publication?: string
	initial_wait_time?: number
	// SSH config fields
	ssh_config?: {
		type?:
			| "no_tunnel"
			| "ssh_key_authentication"
			| "ssh_password_authentication"
		host?: string
		port?: number
		username?: string
		password?: string
		private_key?: string
		passphrase?: string
	}
}): SourceFormConfig => ({
	name: data.name,
	connector: SourceConnector.Postgres,
	version: data.version,
	fields: {
		host: data.host,
		port: data.port.toString(),
		database: data.database,
		username: data.username,
		password: data.password,
		// Handle nested SSL config - supports both formats
		...(data.ssl_mode && {
			ssl: { mode: data.ssl_mode }, // Nested object format (auto-flattened)
		}),
		...(data.max_threads && { max_threads: data.max_threads }),
		...(data.retry_count && { retry_count: data.retry_count }),
		...(data.jdbc_url_params && { jdbc_url_params: data.jdbc_url_params }),
		// Handle nested update_method config
		...(data.update_method === "CDC" && {
			update_method: {
				type: "CDC",
				...(data.replication_slot && {
					replication_slot: data.replication_slot,
				}),
				...(data.publication && { publication: data.publication }),
				...(data.initial_wait_time && {
					initial_wait_time: data.initial_wait_time,
				}),
			},
		}),
		// Handle nested SSH config
		...(data.ssh_config && { ssh_config: data.ssh_config }),
	},
})

/**
 * MongoDB Source Configuration Template
 * Based on the actual MongoDB connector schema
 */
export const createMongoDBSourceConfig = (data: {
	name: string
	hosts: string[]
	database: string
	username: string
	password: string
	version?: string
	// Optional fields from schema
	authdb?: string
	replica_set?: string
	read_preference?: string
	srv?: boolean
	max_threads?: number
	backoff_retry_count?: number
	chunking_strategy?: "Split Vector" | "Timestamp"
}): SourceFormConfig => ({
	name: data.name,
	connector: SourceConnector.MongoDB,
	version: data.version,
	fields: {
		hosts: data.hosts,
		database: data.database,
		authdb: data.authdb || "admin", // Default value from schema
		username: data.username,
		password: data.password,
		...(data.replica_set && { replica_set: data.replica_set }),
		...(data.read_preference && { read_preference: data.read_preference }),
		...(data.srv !== undefined && { srv: data.srv }),
		...(data.max_threads && { max_threads: data.max_threads }),
		...(data.backoff_retry_count && {
			backoff_retry_count: data.backoff_retry_count,
		}),
		...(data.chunking_strategy && {
			chunking_strategy: data.chunking_strategy,
		}),
	},
})

/**
 * MySQL Source Configuration Template
 * Example of how to add a new connector - no CreateSourcePage changes needed!
 */
export const createMySQLSourceConfig = (data: {
	name: string
	host: string
	port: string
	database: string
	username: string
	password: string
	version?: string
	ssl?: boolean
}): SourceFormConfig => ({
	name: data.name,
	connector: SourceConnector.MySQL,
	version: data.version,
	fields: {
		host: data.host,
		port: data.port,
		database: data.database,
		username: data.username,
		password: data.password,
		...(data.ssl !== undefined && { ssl: data.ssl }),
	},
})

/**
 * Oracle Source Configuration Template
 * Another example of adding a new connector
 */
export const createOracleSourceConfig = (data: {
	name: string
	host: string
	port: string
	serviceName: string
	username: string
	password: string
	version?: string
}): SourceFormConfig => ({
	name: data.name,
	connector: SourceConnector.Oracle,
	version: data.version,
	fields: {
		host: data.host,
		port: data.port,
		service_name: data.serviceName,
		username: data.username,
		password: data.password,
	},
})

/**
 * Generic Source Configuration Builder
 * Use this for custom or one-off configurations
 */
export const createGenericSourceConfig = (
	connector: SourceConnector,
	name: string,
	fields: Record<string, any>,
	version?: string,
): SourceFormConfig => ({
	name,
	connector,
	version,
	fields,
})

/**
 * Test data configurations for E2E tests
 */
export const POSTGRES_SOURCE_CONFIG = {
	host: "172.17.0.2",
	database: "postgres",
	username: "postgres",
	password: "secret1234",
	port: "5433",
} as const

export const MONGODB_SOURCE_CONFIG = {
	host: "172.17.0.2",
	database: "test_db",
	username: "admin",
	password: "password",
	port: "27017",
} as const
