import { test, expect } from "../fixtures/auth.fixture"
import { JOB_TEST_CONFIG } from "../setup/test-env"

test.describe("Job End-to-End User Journey", () => {
	test("should complete full job workflow: create source → create destination → create job → sync", async ({
		loginPage,
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		jobsPage,
		createJobPage,
		page,
	}) => {
		const timestamp = Date.now()
		const sourceData = {
			name: `e2e-source-${timestamp}`,
			host: "host.docker.internal",
			database: "postgres",
			username: "postgres",
			password: "secret1234",
			useSSL: false,
			port: "5433",
		}

		const destinationData = {
			name: `e2e-destination-${timestamp}`,
			jdbcUrl: "jdbc:postgresql://host.docker.internal:5432/iceberg",
			jdbcUsername: "iceberg",
			jdbcPassword: "password",
			jdbcDatabase: "iceberg",
			jdbcS3Endpoint: "http://host.docker.internal:9000",
			jdbcS3AccessKey: "admin",
			jdbcS3SecretKey: "password",
			jdbcS3Region: "us-east-1",
			jdbcS3Path: "s3a://warehouse",
			jdbcUsePathStyleForS3: true,
			jdbcUseSSLForS3: false,
		}

		const jobData = {
			sourceName: sourceData.name,
			destinationName: destinationData.name,
			streamName: "postgres_test_table_olake",
			jobName: `e2ejob${timestamp}`,
			frequency: JOB_TEST_CONFIG.frequency,
		}

		// Step 1: Login
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()
		await expect(page).toHaveURL("/jobs")

		// Step 2: Create Source
		await sourcesPage.navigateToSources()
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.clickCreateSource()
		await createSourcePage.expectCreateSourcePageVisible()
		// await createSourcePage.fillMongoDBForm(sourceData)
		await createSourcePage.selectPostgresFillPostgresCreds(sourceData)
		await createSourcePage.clickCreate()
		await createSourcePage.expectTestConnectionModal()
		await createSourcePage.expectSuccessModal()
		await createSourcePage.expectEntitySavedModal()
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(sourceData.name)

		// Step 3: Create Destination
		await destinationsPage.navigateToDestinations()
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.clickCreateDestination()
		await createDestinationPage.expectCreateDestinationPageVisible()
		// await createDestinationPage.fillAmazonS3Form(destinationData)
		await createDestinationPage.fillIcebergJdbcForm(destinationData)
		await createDestinationPage.clickCreate()
		await createDestinationPage.expectTestConnectionModal()
		await createDestinationPage.expectSuccessModal()
		await createDestinationPage.expectEntitySavedModal()
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(destinationData.name)

		// Step 4: Create Job
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
		await jobsPage.clickCreateJob()
		await createJobPage.expectCreateJobPageVisible()
		await createJobPage.fillJobCreationForm(jobData)
		await createJobPage.goToJobsPage()
		await jobsPage.expectJobsPageVisible()
		// await jobsPage.expectJobExists(jobData.jobName)

		// Step 5: Sync Job
		await jobsPage.syncJob(jobData.jobName)
		await expect(page).toHaveURL(/\/jobs\/.*\/history/)

		// Step 6: View Logs and Configurations
		await jobsPage.viewJobLogs()
		await jobsPage.expectLogsCellVisible()
		await jobsPage.viewJobConfigurations()

		// Step 7: Navigate back and verify
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
		await jobsPage.expectJobExists(jobData.jobName)
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
