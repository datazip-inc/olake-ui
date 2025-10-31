import { test, expect } from "../fixtures/auth.fixture"
import {
	S3_TEST_CONFIG,
	VALIDATION_MESSAGES,
	ICEBERG_JDBC_TEST_CONFIG,
} from "../setup/test-env"

test.describe("Create Destination Flow", () => {
	test.beforeEach(async ({ loginPage, destinationsPage }) => {
		// Login first
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()

		// Navigate to destinations and create new destination
		await destinationsPage.navigateToDestinations()
		await destinationsPage.clickCreateDestination()
	})

	test("should display create destination page correctly", async ({
		createDestinationPage,
	}) => {
		await createDestinationPage.expectCreateDestinationPageVisible()
	})

	test("should create Amazon S3 destination successfully", async ({
		createDestinationPage,
		destinationsPage,
	}) => {
		const destinationData = {
			name: `test-s3-${Date.now()}`,
			bucketName: S3_TEST_CONFIG.bucketName,
			region: S3_TEST_CONFIG.region,
			path: S3_TEST_CONFIG.path,
		}

		// Fill the S3 form
		await createDestinationPage.fillAmazonS3Form(destinationData)

		// Create the destination
		await createDestinationPage.clickCreate()

		// Wait for test connection and success modals
		await createDestinationPage.expectTestConnectionModal()
		await createDestinationPage.expectSuccessModal()
		await createDestinationPage.expectEntitySavedModal()

		// Should navigate back to destinations page
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(destinationData.name)
	})

	test("should create Apache Iceberg JDBC Catalogue destination successfully", async ({
		createDestinationPage,
		destinationsPage,
	}) => {
		const destinationData = {
			name: `test-iceberg-jdbc-${Date.now()}`,
			jdbcUrl: ICEBERG_JDBC_TEST_CONFIG.jdbc_url,
			jdbcUsername: ICEBERG_JDBC_TEST_CONFIG.jdbc_username,
			jdbcPassword: ICEBERG_JDBC_TEST_CONFIG.jdbc_password,
			jdbcDatabase: ICEBERG_JDBC_TEST_CONFIG.jdbc_database,
			jdbcS3Endpoint: ICEBERG_JDBC_TEST_CONFIG.jdbc_s3_endpoint,
			jdbcS3AccessKey: ICEBERG_JDBC_TEST_CONFIG.jdbc_s3_access_key,
			jdbcS3SecretKey: ICEBERG_JDBC_TEST_CONFIG.jdbc_s3_secret_key,
			jdbcS3Region: ICEBERG_JDBC_TEST_CONFIG.jdbc_s3_region,
			jdbcS3Path: ICEBERG_JDBC_TEST_CONFIG.iceberg_s3_path,
			jdbcUsePathStyleForS3: ICEBERG_JDBC_TEST_CONFIG.s3_use_path_style,
			jdbcUseSSLForS3: ICEBERG_JDBC_TEST_CONFIG.s3_use_ssl,
		}

		// Fill the S3 form
		await createDestinationPage.fillIcebergJdbcForm(destinationData)

		// Create the destination
		await createDestinationPage.clickCreate()

		// Wait for test connection and success modals
		await createDestinationPage.expectTestConnectionModal()
		await createDestinationPage.expectSuccessModal()
		await createDestinationPage.expectEntitySavedModal()

		// Should navigate back to destinations page
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(destinationData.name)
	})

	test("should show validation error for empty destination name", async ({
		createDestinationPage,
	}) => {
		// Try to create without filling destination name
		await createDestinationPage.clickCreate()
		await createDestinationPage.expectValidationError(
			VALIDATION_MESSAGES.destinationName.required,
		)
	})

	test("should show validation error for missing configuration", async ({
		createDestinationPage,
	}) => {
		await createDestinationPage.fillDestinationName("test-destination")
		await createDestinationPage.clickCreate()

		// Should show validation for missing required fields
		await expect(
			createDestinationPage.page.locator(".ant-form-item-has-error"),
		).toBeVisible()
	})

	test("should allow canceling destination creation", async ({
		createDestinationPage,
		destinationsPage,
	}) => {
		await createDestinationPage.fillDestinationName("test-cancel")
		await createDestinationPage.clickCancel()

		// Should show confirmation modal and navigate back
		await expect(createDestinationPage.page.getByText("Cancel")).toBeVisible()
		await createDestinationPage.page
			.getByRole("button", { name: "Confirm" })
			.click()

		await destinationsPage.expectDestinationsPageVisible()
	})

	test("should fill form step by step correctly", async ({
		createDestinationPage,
	}) => {
		// Test step-by-step form filling to ensure all fields work
		await createDestinationPage.fillDestinationName("step-by-step-test")
		await expect(createDestinationPage.destinationNameInput).toHaveValue(
			"step-by-step-test",
		)

		await createDestinationPage.selectConnector("Amazon S3")
		await expect(
			createDestinationPage.page.locator("text=Amazon S3"),
		).toBeVisible()

		await createDestinationPage.fillS3Configuration({
			bucketName: "test-bucket",
			region: "us-east-1",
			path: "/test",
		})

		await expect(createDestinationPage.bucketNameInput).toHaveValue(
			"test-bucket",
		)
		await expect(createDestinationPage.regionInput).toHaveValue("us-east-1")
		await expect(createDestinationPage.pathInput).toHaveValue("/test")
	})

	test("should handle connector selection", async ({
		createDestinationPage,
	}) => {
		// Test connector dropdown functionality
		await createDestinationPage.selectConnector("Apache Iceberg")

		// The form should adapt to Apache Iceberg schema
		await expect(
			createDestinationPage.page.locator("text=Apache Iceberg"),
		).toBeVisible()
	})

	test("should support setup type selection", async ({
		createDestinationPage,
	}) => {
		// Test new vs existing destination setup
		await createDestinationPage.selectSetupType("existing")
		await expect(createDestinationPage.setupTypeExisting).toBeVisible()

		await createDestinationPage.selectSetupType("new")
		await expect(createDestinationPage.setupTypeNew).toBeVisible()
	})

	test("should validate required S3 configuration fields", async ({
		createDestinationPage,
	}) => {
		await createDestinationPage.fillDestinationName("s3-validation-test")
		await createDestinationPage.selectConnector("Amazon S3")

		// Fill only bucket name, leave region and path empty
		await createDestinationPage.bucketNameInput.fill("test-bucket")
		await createDestinationPage.clickCreate()

		// Should show validation errors for missing fields
		await expect(
			createDestinationPage.page.locator(".ant-form-item-has-error"),
		).toBeVisible()
	})
})
