import { test, expect } from "../fixtures/auth.fixture"
import { MONGODB_TEST_CONFIG } from "../setup/test-env"

test.describe("Edit Source Flow", () => {
	let testSourceName: string

	test.beforeEach(async ({ loginPage, sourcesPage, createSourcePage }) => {
		testSourceName = `test-source-${Date.now()}`

		// Login first
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()

		// Create a test source first
		await sourcesPage.navigateToSources()
		await sourcesPage.clickCreateSource()

		await createSourcePage.fillMongoDBForm({
			name: testSourceName,
			...MONGODB_TEST_CONFIG,
		})

		await createSourcePage.clickCreate()

		// Wait for source creation to complete
		await createSourcePage.expectEntitySavedModal()
		await sourcesPage.expectSourcesPageVisible()
	})

	test("should display edit source page correctly", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.expectEditSourcePageVisible()
	})

	test("should edit source name successfully", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		const newSourceName = `${testSourceName}-edited`

		// Navigate to edit page
		await sourcesPage.editSource(testSourceName)

		// Update source name
		await editSourcePage.updateSourceName(newSourceName)
		await editSourcePage.clickSaveChanges()
		await editSourcePage.clickConfirm()

		// Should show success and navigate back
		await editSourcePage.expectSuccessMessage()
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(newSourceName)
	})

	test("should view associated jobs", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.viewAssociatedJobs()

		// Should display associated jobs section
		await expect(editSourcePage.associatedJobsButton).toBeVisible()
	})

	test("should cancel edit without saving changes", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		const originalName = testSourceName

		await sourcesPage.editSource(testSourceName)
		await editSourcePage.updateSourceName(`${testSourceName}-cancelled`)
		await editSourcePage.clickCancel()

		// Should navigate back without saving
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(originalName)
	})

	test("should test connection from edit page", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.testConnection()

		// Should show test connection modal
		await editSourcePage.expectSuccessMessage()
	})

	test("should show validation error for empty source name", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.updateSourceName("")
		await editSourcePage.clickSaveChanges()

		await editSourcePage.expectValidationError("Source name is required")
	})

	test("should navigate back to sources page", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.goBackToSources()

		await sourcesPage.expectSourcesPageVisible()
	})

	test("should complete full edit flow with confirmation", async ({
		sourcesPage,
		editSourcePage,
	}) => {
		const updatedName = `${testSourceName}-full-flow`

		// Navigate to edit
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.expectEditSourcePageVisible()

		// View associated jobs first
		await editSourcePage.viewAssociatedJobs()

		// Make changes
		await editSourcePage.updateSourceName(updatedName)

		// Save changes
		await editSourcePage.clickSaveChanges()
		await editSourcePage.expectConfirmationDialog()
		await editSourcePage.clickConfirm()

		// Verify success
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(updatedName)
	})

	test("should handle navigation during edit flow", async ({
		sourcesPage,
		editSourcePage,
		page,
	}) => {
		await sourcesPage.editSource(testSourceName)

		// Test different navigation options
		await editSourcePage.goBackToSources()
		await sourcesPage.expectSourcesPageVisible()

		// Navigate back to edit
		await sourcesPage.editSource(testSourceName)
		await editSourcePage.expectEditSourcePageVisible()

		// Verify URL
		await expect(page.url()).toContain("/sources/")
	})
})
