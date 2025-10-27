// Test environment configuration for Olake UI
// Replace these values with your actual test credentials in a real environment

type TimeoutConfig = {
	short: number
	medium: number
	long: number
}

type TestFeatures = {
	auth: boolean
	cleanup: boolean
	parallel: boolean
}

type TestAdminUser = {
	email: string
	password: string
}

type TestConfig = {
	baseUrl: string
	adminUser: TestAdminUser
	timeouts: TimeoutConfig
	features: TestFeatures
}

// Validation Messages
export const VALIDATION_MESSAGES = {
	required: "This field is required",
	invalidEmail: "Please enter a valid email address",
	minLength: (min: number): string => `Must be at least ${min} characters`,
	maxLength: (max: number): string => `Must be at most ${max} characters`,
	invalidUrl: "Please enter a valid URL",
	invalidNumber: "Please enter a valid number",
	invalidDate: "Please enter a valid date",
	passwordMismatch: "Passwords do not match",
}

// MongoDB Configuration
export const MONGODB_TEST_CONFIG = {
	connectionString: "mongodb://localhost:27017/test-olake",
	dbName: "test-olake",
}

// S3 Configuration
export const S3_TEST_CONFIG = {
	accessKeyId: "test-access-key",
	secretAccessKey: "test-secret-key",
	region: "us-east-1",
	bucket: "test-bucket",
}

// Iceberg JDBC Configuration
export const ICEBERG_JDBC_TEST_CONFIG = {
	jdbcUrl: "jdbc:trino://localhost:8080",
	username: "test-user",
	password: "test-password",
	catalog: "iceberg",
	schema: "test_schema",
}

// Job Test Configuration
export const JOB_TEST_CONFIG = {
	defaultSyncMode: "full_refresh" as const,
	defaultDestinationSchema: "public",
}

// Environment variables with fallbacks
const env = {
	TEST_BASE_URL: "http://localhost:5173",
	TEST_ADMIN_EMAIL: "admin@example.com",
	TEST_ADMIN_PASSWORD: "testpassword123",
	...process.env,
}

// Base Test Configuration
export const TEST_CONFIG: TestConfig = {
	baseUrl: env.TEST_BASE_URL,
	adminUser: {
		email: env.TEST_ADMIN_EMAIL,
		password: env.TEST_ADMIN_PASSWORD,
	},
	// Timeout configurations
	timeouts: {
		short: 5000,
		medium: 15000,
		long: 30000,
	},
	// Enable/disable test features
	features: {
		auth: true,
		cleanup: true,
		parallel: false,
	},
}
