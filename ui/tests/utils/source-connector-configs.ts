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
import { SERVICE_HOST } from "./constants"
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
 * The Postgres source of the integration test (server/tests/sources/postgres/docker-compose.yml)
 */
export const POSTGRES_SOURCE_CONFIG = {
	host: SERVICE_HOST,
	database: "postgres",
	username: "postgres",
	password: "secret1234",
	port: "5433",
} as const
