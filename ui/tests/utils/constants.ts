/**
 * Test Data Constants
 *
 * Centralized test configuration and data for all E2E tests.
 * This file contains credentials, connection strings, and test configurations
 * that are used across multiple test files.
 */

// Job Configurations
export const JOB_CONFIG = {
	streamName: "postgres_test_table_olake",
	frequency: "Every Week",
} as const

// Validation Messages
export const VALIDATION_MESSAGES = {
	username: {
		required: "Please input your username!",
		minLength: "Username must be at least 3 characters",
	},
	password: {
		required: "Please input your password!",
		minLength: "Password must be at least 6 characters",
	},
	sourceName: {
		required: "Source name is required",
	},
	destinationName: {
		required: "Destination name is required",
	},
	jobName: {
		required: "Job name is required",
	},
} as const

// URLs
export const URLS = {
	login: "/login",
	jobs: "/jobs",
	sources: "/sources",
	destinations: "/destinations",
} as const

export const LOGIN_CREDENTIALS = {
	admin: {
		username: "admin",
		password: "password",
	},
	invalid: {
		username: "invalid",
		password: "invalid",
	},
} as const

export const DESTINATION_CONNECTOR_TEST_ID_MAP: Record<string, string> = {
	"Apache Iceberg": "connector-option-iceberg",
	"Amazon S3": "connector-option-s3",
}

export const SOURCE_CONNECTOR_TEST_ID_MAP: Record<string, string> = {
	MongoDB: "connector-option-mongodb",
	Postgres: "connector-option-postgres",
	MySQL: "connector-option-mysql",
	Oracle: "connector-option-oracle",
	Kafka: "connector-option-kafka",
	S3: "connector-option-s3",
}
