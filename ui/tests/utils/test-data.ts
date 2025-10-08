/**
 * Test Data Constants
 *
 * Centralized test configuration and data for all E2E tests.
 * This file contains credentials, connection strings, and test configurations
 * that are used across multiple test files.
 */

// Authentication
export const TEST_CREDENTIALS = {
	username: "admin",
	password: "password",
} as const

// Source Configurations
export const POSTGRES_SOURCE_CONFIG = {
	host: "172.17.0.2",
	database: "postgres",
	username: "postgres",
	password: "secret1234",
	useSSL: false,
	port: "5433",
} as const

export const MONGODB_TEST_CONFIG = {
	host: "172.17.0.2",
	database: "test_db",
	username: "admin",
	password: "password",
	useSSL: false,
	port: "27017",
} as const

// Destination Configurations
export const ICEBERG_DESTINATION_CONFIG = {
	jdbcUrl: "jdbc:postgresql://172.17.0.2:5432/iceberg",
	jdbcUsername: "iceberg",
	jdbcPassword: "password",
	jdbcDatabase: "olake_iceberg",
	jdbcS3Endpoint: "http://172.17.0.2:9000",
	jdbcS3AccessKey: "admin",
	jdbcS3SecretKey: "password",
	jdbcS3Region: "us-east-1",
	jdbcS3Path: "s3a://warehouse",
	jdbcUsePathStyleForS3: true,
	jdbcUseSSLForS3: false,
} as const

export const S3_TEST_CONFIG = {
	bucketName: "test-bucket",
	region: "us-east-1",
	path: "s3://test-bucket/data",
	accessKey: "admin",
	secretKey: "password",
} as const

// Job Configurations
export const JOB_CONFIG = {
	streamName: "postgres_test_table_olake",
	frequency: "Every Week",
} as const

// Validation Messages
export const VALIDATION_MESSAGES = {
	username: {
		required: "Username is required",
		minLength: "Username must be at least 3 characters",
	},
	password: {
		required: "Password is required",
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
