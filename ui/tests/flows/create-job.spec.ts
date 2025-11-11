import { test, expect } from "../fixtures/auth.fixture"
import {
	MONGODB_TEST_CONFIG,
	S3_TEST_CONFIG,
	JOB_TEST_CONFIG,
	VALIDATION_MESSAGES,
} from "../setup/test-env"

test.describe("Create Job Flow", () => {
	let testSourceName: string
	let testDestinationName: string

	test.beforeEach(
		async ({
			loginPage,
			sourcesPage,
			createSourcePage,
			destinationsPage,
			createDestinationPage,
			jobsPage,
		}) => {
			testSourceName = `test-source-${Date.now()}`
			testDestinationName = `test-destination-${Date.now()}`

			// Login first
			await loginPage.goto()
			await loginPage.login("admin", "password")
			await loginPage.waitForLogin()

			// Create a test source
			await sourcesPage.navigateToSources()
			await sourcesPage.clickCreateSource()
			await createSourcePage.fillMongoDBForm({
				name: testSourceName,
				host: MONGODB_TEST_CONFIG.host,
				database: MONGODB_TEST_CONFIG.database,
				username: MONGODB_TEST_CONFIG.username,
				password: MONGODB_TEST_CONFIG.password,
				useSSL: MONGODB_TEST_CONFIG.useSSL,
			})
			await createSourcePage.clickCreate()
			await createSourcePage.expectEntitySavedModal()

			// Create a test destination
			await destinationsPage.navigateToDestinations()
			await destinationsPage.clickCreateDestination()
			await createDestinationPage.fillAmazonS3Form({
				name: testDestinationName,
				bucketName: S3_TEST_CONFIG.bucketName,
				region: S3_TEST_CONFIG.region,
				path: S3_TEST_CONFIG.path,
			})
			await createDestinationPage.clickCreate()
			await createDestinationPage.expectEntitySavedModal()

			// Navigate to jobs page to start job creation
			await jobsPage.navigateToJobs()
			await jobsPage.clickCreateJob()
		},
	)

	test("should display create job page correctly", async ({
		createJobPage,
	}) => {
		await createJobPage.expectCreateJobPageVisible()
	})

	test("should create job successfully with existing source and destination", async ({
		createJobPage,
		jobsPage,
	}) => {
		const jobData = {
			sourceName: testSourceName,
			destinationName: testDestinationName,
			streamName: JOB_TEST_CONFIG.streamName,
			jobName: `test-job-${Date.now()}`,
			frequency: JOB_TEST_CONFIG.frequency,
		}

		// Fill the job creation form
		await createJobPage.fillJobCreationForm(jobData)

		// Navigate to jobs page
		await createJobPage.goToJobsPage()

		// Verify job was created
		await jobsPage.expectJobsPageVisible()
		await jobsPage.expectJobExists(jobData.jobName)
	})

	test("should validate required job name", async ({ createJobPage }) => {
		// Go through the steps without filling job name
		await createJobPage.selectExistingSource(testSourceName)
		await createJobPage.selectExistingDestination(testDestinationName)
		await createJobPage.configureStreams(JOB_TEST_CONFIG.streamName)

		// Try to create job without name
		await createJobPage.createJobButton.click()
		await createJobPage.expectValidationError(
			VALIDATION_MESSAGES.jobName.required,
		)
	})

	test("should allow step-by-step job configuration", async ({
		createJobPage,
	}) => {
		// Step 1: Source selection
		await createJobPage.selectExistingSource(testSourceName)

		// Step 2: Destination selection
		await createJobPage.selectExistingDestination(testDestinationName)

		// Step 3: Stream configuration
		await createJobPage.configureStreams(JOB_TEST_CONFIG.streamName)

		// Step 4: Job settings
		const testJobName = "step-by-step-test"
		await createJobPage.configureJobSettings(testJobName, "Every Week")

		// Should complete successfully
		await createJobPage.goToJobsPage()
	})

	test("should handle stream configuration options", async ({
		createJobPage,
	}) => {
		await createJobPage.selectExistingSource(testSourceName)
		await createJobPage.selectExistingDestination(testDestinationName)

		// Test sync all checkbox
		await expect(createJobPage.syncAllCheckbox).toBeVisible()
		await createJobPage.syncAllCheckbox.uncheck()

		// Test stream selection
		await createJobPage.page
			.getByRole("button", { name: JOB_TEST_CONFIG.streamName })
			.getByLabel("")
			.check()

		// Test sync mode selection
		await createJobPage.fullRefreshIncrementalRadio.check()
		await expect(createJobPage.fullRefreshIncrementalRadio).toBeChecked()

		// Test stream switch
		await createJobPage.page.getByRole("switch").first().click()

		await createJobPage.nextButton.click()
	})

	test("should support different frequency options", async ({
		createJobPage,
	}) => {
		await createJobPage.selectExistingSource(testSourceName)
		await createJobPage.selectExistingDestination(testDestinationName)
		await createJobPage.configureStreams(JOB_TEST_CONFIG.streamName)

		// Test frequency selection
		await createJobPage.jobNameInput.fill("frequency-test")
		await createJobPage.frequencyDropdown.click()
		await createJobPage.page.getByText("Every Hour").click()

		await expect(createJobPage.page.getByText("Every Hour")).toBeVisible()
	})

	test("should handle navigation between steps", async ({ createJobPage }) => {
		// Test forward navigation
		await createJobPage.selectExistingSource(testSourceName)
		await createJobPage.selectExistingDestination(testDestinationName)
		await createJobPage.configureStreams(JOB_TEST_CONFIG.streamName)

		// Should be on the final step
		await expect(createJobPage.jobNameInput).toBeVisible()
		await expect(createJobPage.createJobButton).toBeVisible()
	})
})
