import { test, expect } from "../fixtures/auth.fixture"
import {
	MONGODB_TEST_CONFIG,
	S3_TEST_CONFIG,
	JOB_TEST_CONFIG,
} from "../setup/test-env"

test.describe("Job Sync Flow", () => {
	let testJobName: string
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
			createJobPage,
		}) => {
			testJobName = `test-job-${Date.now()}`
			testSourceName = `test-source-${Date.now()}`
			testDestinationName = `test-destination-${Date.now()}`

			// Login first
			await loginPage.goto()
			await loginPage.login("admin", "password")
			await loginPage.waitForLogin()

			// Create prerequisites (source and destination)
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

			// Create a job
			await jobsPage.navigateToJobs()
			await jobsPage.clickCreateJob()
			await createJobPage.fillJobCreationForm({
				sourceName: testSourceName,
				destinationName: testDestinationName,
				streamName: JOB_TEST_CONFIG.streamName,
				jobName: testJobName,
				frequency: JOB_TEST_CONFIG.frequency,
			})
			await createJobPage.goToJobsPage()
			await jobsPage.expectJobExists(testJobName)
		},
	)

	test("should sync job successfully", async ({ jobsPage }) => {
		// Trigger job sync
		await jobsPage.syncJob(testJobName)

		// Should navigate to job history page
		await expect(jobsPage.page).toHaveURL(/\/jobs\/.*\/history/)
	})

	test("should view job logs after sync", async ({ jobsPage, page }) => {
		// Sync the job
		await jobsPage.syncJob(testJobName)

		// Wait a moment for sync to start
		await page.waitForTimeout(2000)

		// View logs
		await jobsPage.viewJobLogs()

		// Should see logs content
		await jobsPage.expectLogsCellVisible()
	})

	test("should view job configurations", async ({ jobsPage }) => {
		await jobsPage.syncJob(testJobName)

		// View job configurations
		await jobsPage.viewJobConfigurations()

		// Should navigate to job configuration page
		await expect(jobsPage.page).toHaveURL(/\/jobs\/.*/)
	})

	test("should navigate back to jobs list", async ({ jobsPage, page }) => {
		await jobsPage.syncJob(testJobName)
		await jobsPage.viewJobLogs()

		// Navigate back to jobs
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
	})

	test("should handle multiple sync operations", async ({ jobsPage, page }) => {
		// First sync
		await jobsPage.syncJob(testJobName)
		await page.waitForTimeout(1000)

		// Navigate back to jobs
		await jobsPage.navigateToJobs()

		// Second sync
		await jobsPage.syncJob(testJobName)

		// Should handle multiple syncs gracefully
		await expect(page).toHaveURL(/\/jobs\/.*\/history/)
	})

	test("should display job details after sync", async ({ jobsPage, page }) => {
		await jobsPage.syncJob(testJobName)

		// Check various elements are clickable/visible
		const buttons = page.getByRole("button").filter({ hasText: /^$/ })
		await expect(buttons.first()).toBeVisible()

		// View logs functionality
		await jobsPage.viewJobLogs()
		await jobsPage.expectLogsCellVisible()

		// View configurations
		await jobsPage.viewJobConfigurations()
		await expect(page).toHaveURL(/\/jobs\/.*/)
	})

	test("should support job sync workflow with navigation", async ({
		jobsPage,
		page,
	}) => {
		// Complete sync workflow as described in the original test
		await jobsPage.syncJob(testJobName)

		// Multiple button clicks (as per original test pattern)
		const buttonSelector = page.getByRole("button").filter({ hasText: /^$/ })
		for (let i = 0; i < 3; i++) {
			await buttonSelector.nth(2).click()
			await page.waitForTimeout(500)
		}

		// View logs
		await jobsPage.viewJobLogs()

		// More button interactions
		for (let i = 0; i < 3; i++) {
			await buttonSelector.nth(2).click()
			await page.waitForTimeout(500)
		}

		// Check logs content
		await jobsPage.expectLogsCellVisible()

		// View job configurations
		await jobsPage.viewJobConfigurations()

		// Navigate back to jobs
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
	})
})
