import { test, expect } from "../fixtures/auth.fixture"
import { MONGODB_TEST_CONFIG } from "../setup/test-env"

test.describe("End-to-End User Journey", () => {
	test("should complete full user flow: login → create source → edit source", async ({
		loginPage,
		sourcesPage,
		createSourcePage,
		editSourcePage,
		page,
	}) => {
		const sourceData = {
			name: `e2e-test-${Date.now()}`,
			...MONGODB_TEST_CONFIG,
		}

		// Step 1: Login
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()
		await expect(page).toHaveURL("/jobs")

		// Step 2: Navigate to Sources
		await sourcesPage.navigateToSources()
		await sourcesPage.expectSourcesPageVisible()

		// Step 3: Create Source
		await sourcesPage.clickCreateSource()
		await createSourcePage.expectCreateSourcePageVisible()

		await createSourcePage.fillMongoDBForm(sourceData)
		await createSourcePage.clickCreate()

		// Wait for source creation success
		await createSourcePage.expectTestConnectionModal()
		await createSourcePage.expectSuccessModal()
		await createSourcePage.expectEntitySavedModal()

		// Should be back on sources page
		await sourcesPage.expectSourcesPageVisible()
		// await sourcesPage.expectSourceExists(sourceData.name)

		// Step 4: Edit the created source
		await sourcesPage.editSource(sourceData.name)
		await editSourcePage.expectEditSourcePageVisible()

		// View associated jobs
		await editSourcePage.viewAssociatedJobs()

		// Update source name
		const updatedName = `${sourceData.name}-edited`
		await editSourcePage.updateSourceName(updatedName)

		// Save changes
		await editSourcePage.clickSaveChanges()
		await editSourcePage.clickConfirm()

		// Step 5: Verify final state
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(updatedName)
		await sourcesPage.expectSourceNotExists(sourceData.name)

		// Navigate back to sources to verify persistence
		await page.goto("/sources")
		await sourcesPage.expectSourceExists(updatedName)
	})

	test("should handle error scenarios gracefully", async ({
		loginPage,
		sourcesPage,
		createSourcePage,
	}) => {
		// Login
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()

		// Try to create source with invalid data
		await sourcesPage.navigateToSources()
		await sourcesPage.clickCreateSource()

		// Submit without filling required fields
		await createSourcePage.clickCreate()
		await createSourcePage.expectValidationError("Source name is required")

		// Fill minimal data and try again
		await createSourcePage.fillSourceName("error-test")
		await createSourcePage.clickCreate()

		// Should show validation for missing fields
		await expect(
			createSourcePage.page.locator(".ant-form-item-has-error"),
		).toBeVisible()

		// Cancel and verify we can navigate back
		await createSourcePage.clickCancel()
		await createSourcePage.page.getByRole("button", { name: "Confirm" }).click()
		await sourcesPage.expectSourcesPageVisible()
	})

	test("should support keyboard navigation throughout the flow", async ({
		loginPage,
		sourcesPage,
		createSourcePage,
		page,
	}) => {
		// Login with keyboard
		await loginPage.goto()
		await page.keyboard.press("Tab") // Focus username
		await page.keyboard.type("admin")
		await page.keyboard.press("Tab") // Focus password
		await page.keyboard.type("password")
		await page.keyboard.press("Enter") // Submit

		await loginPage.waitForLogin()

		// Navigate using keyboard
		await sourcesPage.navigateToSources()
		await sourcesPage.clickCreateSource()

		// Fill form using keyboard navigation
		await page.keyboard.press("Tab") // Navigate to source name
		await page.keyboard.type("keyboard-test")

		// Can continue with Tab navigation through the form
		await expect(createSourcePage.sourceNameInput).toHaveValue("keyboard-test")
	})
})
