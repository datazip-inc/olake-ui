import { testAuthenticated as test, expect } from "../fixtures"
import {
	createPostgresSourceConfig,
	POSTGRES_SOURCE_CONFIG,
	createIcebergJdbcConfig,
	ICEBERG_JDBC_CONFIG,
	JOB_CONFIG,
	TestDataBuilder,
	verifyEntityCreationSuccessModal,
} from "../utils"

test.describe("Job End-to-End User Journey", () => {
	test("should complete full job workflow: create source → create destination → create job → sync", async ({
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		jobsPage,
		createJobPage,
		page,
	}) => {
		await jobsPage.goto()
		await expect(page).toHaveURL("/jobs")

		const SOURCE_CONNECTOR = "postgres"
		const DEST_CONNECTOR = "iceberg"
		const CATALOG_TYPE = "jdbc"

		const sourceName = TestDataBuilder.getUniqueSourceName(SOURCE_CONNECTOR)
		const destinationName =
			TestDataBuilder.getUniqueDestinationName(DEST_CONNECTOR)
		const jobName = TestDataBuilder.getUniqueJobName(
			SOURCE_CONNECTOR,
			DEST_CONNECTOR,
			CATALOG_TYPE,
		)

		// Step 1: Create PostgreSQL Source
		await sourcesPage.navigateToSources()
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.clickCreateSource()
		await createSourcePage.expectCreateSourcePageVisible()

		const sourceConfig = createPostgresSourceConfig({
			name: sourceName,
			...POSTGRES_SOURCE_CONFIG,
		})

		await createSourcePage.fillSourceForm(sourceConfig)
		await createSourcePage.clickCreate()
		await verifyEntityCreationSuccessModal(createSourcePage)
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(sourceName)

		// Step 2: Create Iceberg JDBC Destination
		await destinationsPage.navigateToDestinations()
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.clickCreateDestination()
		await createDestinationPage.expectCreateDestinationPageVisible()

		const destinationConfig = createIcebergJdbcConfig({
			name: destinationName,
			...ICEBERG_JDBC_CONFIG,
		})

		await createDestinationPage.fillDestinationForm(destinationConfig)
		await createDestinationPage.clickCreate()
		await verifyEntityCreationSuccessModal(createDestinationPage)
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(destinationName)

		// Step 3: Create Job
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
		await jobsPage.clickCreateJob()
		await createJobPage.expectCreateJobPageVisible()

		await createJobPage.fillJobCreationForm({
			sourceName,
			destinationName,
			jobName,
			streamName: JOB_CONFIG.streamName,
			frequency: JOB_CONFIG.frequency,
		})

		await createJobPage.goToJobsPage()
		await jobsPage.expectJobsPageVisible()

		// Step 4: Sync Job and Verify
		await jobsPage.syncJob(jobName)
		await expect(page).toHaveURL(/\/jobs\/.*\/history/)

		// Step 5: Verify Job Details
		await jobsPage.viewJobLogs()
		await jobsPage.viewJobConfigurations()

		// Step 6: Verify Job in List
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
		await jobsPage.expectJobExists(jobName)
	})

	// test("should handle error scenarios in job workflow", async ({
	// 	loginPage,
	// 	jobsPage,
	// 	createJobPage,
	// }) => {
	// 	// Login
	// 	await loginPage.goto()
	// 	await loginPage.login("admin", "password")
	// 	await loginPage.waitForLogin()

	// 	// Try to create job without prerequisites
	// 	await jobsPage.navigateToJobs()
	// 	await jobsPage.clickCreateJob()

	// 	// Should be able to access the form but validation will fail
	// 	await createJobPage.expectCreateJobPageVisible()

	// 	// Try to proceed without selecting source/destination
	// 	// This would typically show validation errors or prevent progression
	// 	await expect(createJobPage.nextButton).toBeVisible()
	// })

	// test("should support job creation with different configurations", async ({
	// 	loginPage,
	// 	sourcesPage,
	// 	createSourcePage,
	// 	destinationsPage,
	// 	createDestinationPage,
	// 	jobsPage,
	// 	createJobPage,
	// }) => {
	// 	const timestamp = Date.now()

	// 	// Login
	// 	await loginPage.goto()
	// 	await loginPage.login("admin", "password")
	// 	await loginPage.waitForLogin()

	// 	// Create minimal prerequisites
	// 	const sourceData = {
	// 		name: `config-test-source-${timestamp}`,
	// 		host: MONGODB_TEST_CONFIG.host,
	// 		database: MONGODB_TEST_CONFIG.database,
	// 		username: MONGODB_TEST_CONFIG.username,
	// 		password: MONGODB_TEST_CONFIG.password,
	// 		useSSL: MONGODB_TEST_CONFIG.useSSL,
	// 	}

	// 	const destinationData = {
	// 		name: `config-test-destination-${timestamp}`,
	// 		bucketName: S3_TEST_CONFIG.bucketName,
	// 		region: S3_TEST_CONFIG.region,
	// 		path: S3_TEST_CONFIG.path,
	// 	}

	// 	// Create source
	// 	await sourcesPage.navigateToSources()
	// 	await sourcesPage.clickCreateSource()
	// 	await createSourcePage.fillMongoDBForm(sourceData)
	// 	await createSourcePage.clickCreate()
	// 	await createSourcePage.expectEntitySavedModal()

	// 	// Create destination
	// 	await destinationsPage.navigateToDestinations()
	// 	await destinationsPage.clickCreateDestination()
	// 	await createDestinationPage.fillAmazonS3Form(destinationData)
	// 	await createDestinationPage.clickCreate()
	// 	await createDestinationPage.expectEntitySavedModal()

	// 	// Test different job configurations
	// 	await jobsPage.navigateToJobs()
	// 	await jobsPage.clickCreateJob()

	// 	// Test frequency options
	// 	await createJobPage.selectExistingSource(sourceData.name)
	// 	await createJobPage.selectExistingDestination(destinationData.name)
	// 	await createJobPage.configureStreams(JOB_TEST_CONFIG.streamName)

	// 	// Test different frequency settings
	// 	await createJobPage.jobNameInput.fill("config-test-job")
	// 	await createJobPage.frequencyDropdown.click()
	// 	await createJobPage.page.getByText("Every Hour").click()

	// 	await createJobPage.createJobButton.click()
	// 	await createJobPage.goToJobsPage()
	// })

	// test("should support keyboard navigation throughout job workflow", async ({
	// 	loginPage,
	// 	jobsPage,
	// 	createJobPage,
	// 	page,
	// }) => {
	// 	// Login with keyboard
	// 	await loginPage.goto()
	// 	await page.keyboard.press("Tab")
	// 	await page.keyboard.type("admin")
	// 	await page.keyboard.press("Tab")
	// 	await page.keyboard.type("password")
	// 	await page.keyboard.press("Enter")

	// 	await loginPage.waitForLogin()

	// 	// Navigate to job creation
	// 	await jobsPage.navigateToJobs()
	// 	await jobsPage.clickCreateJob()

	// 	// Test keyboard navigation in job form
	// 	await page.keyboard.press("Tab")
	// 	await expect(createJobPage.useExistingSourceRadio).toBeFocused()

	// 	// Can continue with Tab navigation through the form
	// 	await createJobPage.expectCreateJobPageVisible()
	// })
})
