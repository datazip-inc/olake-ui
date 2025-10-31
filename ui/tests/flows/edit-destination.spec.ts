import { test, expect } from "../fixtures/auth.fixture"
import { S3_TEST_CONFIG, VALIDATION_MESSAGES } from "../setup/test-env"

test.describe("Edit Destination Flow", () => {
	let testDestinationName: string

	test.beforeEach(
		async ({ loginPage, destinationsPage, createDestinationPage }) => {
			testDestinationName = `test-destination-${Date.now()}`

			// Login first
			await loginPage.goto()
			await loginPage.login("admin", "password")
			await loginPage.waitForLogin()

			// Create a test destination first
			await destinationsPage.navigateToDestinations()
			await destinationsPage.clickCreateDestination()

			await createDestinationPage.fillAmazonS3Form({
				name: testDestinationName,
				bucketName: S3_TEST_CONFIG.bucketName,
				region: S3_TEST_CONFIG.region,
				path: S3_TEST_CONFIG.path,
			})

			await createDestinationPage.clickCreate()

			// Wait for destination creation to complete
			await createDestinationPage.expectEntitySavedModal()
			await destinationsPage.expectDestinationsPageVisible()
		},
	)

	test("should display edit destination page correctly", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.expectEditDestinationPageVisible()
	})

	test("should edit destination name successfully", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		const newDestinationName = `${testDestinationName}-edited`

		// Navigate to edit page
		await destinationsPage.editDestination(testDestinationName)

		// Update destination name
		await editDestinationPage.updateDestinationName(newDestinationName)
		await editDestinationPage.clickSaveChanges()
		await editDestinationPage.clickConfirm()

		// Should show success and navigate back
		await editDestinationPage.expectSuccessMessage()
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(newDestinationName)
	})

	test("should view associated jobs", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.viewAssociatedJobs()

		// Should display associated jobs section
		await expect(editDestinationPage.associatedJobsButton).toBeVisible()
	})

	test("should view config", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.viewConfig()

		// Should display config section
		await expect(editDestinationPage.configButton).toBeVisible()
	})

	test("should cancel edit without saving changes", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		const originalName = testDestinationName

		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.updateDestinationName(
			`${testDestinationName}-cancelled`,
		)
		await editDestinationPage.clickCancel()

		// Should navigate back without saving
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(originalName)
	})

	test("should test connection from edit page", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.testConnection()

		// Should show test connection modal
		await editDestinationPage.expectSuccessMessage()
	})

	test("should show validation error for empty destination name", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.updateDestinationName("")
		await editDestinationPage.clickSaveChanges()

		await editDestinationPage.expectValidationError(
			VALIDATION_MESSAGES.destinationName.required,
		)
	})

	test("should navigate back to destinations page", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.goBackToDestinations()

		await destinationsPage.expectDestinationsPageVisible()
	})

	test("should complete full edit flow with config and jobs", async ({
		destinationsPage,
		editDestinationPage,
	}) => {
		const updatedName = `${testDestinationName}-full-flow`

		// Navigate to edit
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.expectEditDestinationPageVisible()

		// View associated jobs first
		await editDestinationPage.viewAssociatedJobs()

		// View config
		await editDestinationPage.viewConfig()

		// Make changes
		await editDestinationPage.updateDestinationName(updatedName)

		// Save changes
		await editDestinationPage.clickSaveChanges()
		await editDestinationPage.expectConfirmationDialog()
		await editDestinationPage.clickConfirm()

		// Verify success
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(updatedName)
	})

	test("should handle navigation during edit flow", async ({
		destinationsPage,
		editDestinationPage,
		page,
	}) => {
		await destinationsPage.editDestination(testDestinationName)

		// Test different navigation options
		await editDestinationPage.goBackToDestinations()
		await destinationsPage.expectDestinationsPageVisible()

		// Navigate back to edit
		await destinationsPage.editDestination(testDestinationName)
		await editDestinationPage.expectEditDestinationPageVisible()

		// Verify URL
		await expect(page.url()).toContain("/destinations/")
	})
})
