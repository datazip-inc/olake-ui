/**
 * Test Data Constants
 *
 * Centralized test configuration and data for all E2E tests.
 * This file contains credentials, connection strings, and test configurations
 * that are used across multiple test files.
 */

// Address the sync driver uses to reach the test's source and destination, which publish their
// ports on the machine running Docker. The Go integration test passes it as E2E_SERVICE_HOST;
// 172.17.0.1, the usual gateway of Docker's default bridge, is the address of that machine as
// the driver containers see it.
export const SERVICE_HOST = process.env.E2E_SERVICE_HOST?.trim() || "172.17.0.1"

export const LOGIN_CREDENTIALS = {
	admin: {
		username: "admin",
		password: "password",
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
