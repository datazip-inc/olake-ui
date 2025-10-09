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
