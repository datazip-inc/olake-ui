import { test, expect } from "../fixtures/auth.fixture"
import { S3_TEST_CONFIG, VALIDATION_MESSAGES } from "../setup/test-env"

test.describe("Destination End-to-End User Journey", () => {
	test("should complete full destination flow: login → create destination → edit destination", async ({
		loginPage,
		destinationsPage,
		createDestinationPage,
		editDestinationPage,
		page,
	}) => {
		const destinationData = {
			name: `e2e-destination-${Date.now()}`,
			bucketName: S3_TEST_CONFIG.bucketName,
			region: S3_TEST_CONFIG.region,
			path: S3_TEST_CONFIG.path,
		}

		// Step 1: Login
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()
		await expect(page).toHaveURL("/jobs")

		// Step 2: Navigate to Destinations
		await destinationsPage.navigateToDestinations()
		await destinationsPage.expectDestinationsPageVisible()

		// Step 3: Create Destination
		await destinationsPage.clickCreateDestination()
		await createDestinationPage.expectCreateDestinationPageVisible()

		await createDestinationPage.fillAmazonS3Form(destinationData)
		await createDestinationPage.clickCreate()

		// Wait for destination creation success
		await createDestinationPage.expectTestConnectionModal()
		await createDestinationPage.expectSuccessModal()
		await createDestinationPage.expectEntitySavedModal()

		// Should be back on destinations page
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(destinationData.name)

		// Step 4: Edit the created destination
		await destinationsPage.editDestination(destinationData.name)
		await editDestinationPage.expectEditDestinationPageVisible()

		// View associated jobs
		await editDestinationPage.viewAssociatedJobs()

		// View config
		await editDestinationPage.viewConfig()

		// Update destination name
		const updatedName = `${destinationData.name}-edited`
		await editDestinationPage.updateDestinationName(updatedName)

		// Save changes
		await editDestinationPage.clickSaveChanges()
		await editDestinationPage.clickConfirm()

		// Step 5: Verify final state
		await destinationsPage.expectDestinationsPageVisible()
		await destinationsPage.expectDestinationExists(updatedName)
		await destinationsPage.expectDestinationNotExists(destinationData.name)

		// Navigate back to destinations to verify persistence
		await page.goto("/destinations")
		await destinationsPage.expectDestinationExists(updatedName)
	})

	test("should handle destination error scenarios gracefully", async ({
		loginPage,
		destinationsPage,
		createDestinationPage,
	}) => {
		// Login
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()

		// Try to create destination with invalid data
		await destinationsPage.navigateToDestinations()
		await destinationsPage.clickCreateDestination()

		// Submit without filling required fields
		await createDestinationPage.clickCreate()
		await createDestinationPage.expectValidationError(
			VALIDATION_MESSAGES.destinationName.required,
		)

		// Fill minimal data and try again
		await createDestinationPage.fillDestinationName("error-test")
		await createDestinationPage.clickCreate()

		// Should show validation for missing fields
		await expect(
			createDestinationPage.page.locator(".ant-form-item-has-error"),
		).toBeVisible()

		// Cancel and verify we can navigate back
		await createDestinationPage.clickCancel()
		await createDestinationPage.page
			.getByRole("button", { name: "Confirm" })
			.click()
		await destinationsPage.expectDestinationsPageVisible()
	})

	test("should support different connector types", async ({
		loginPage,
		destinationsPage,
		createDestinationPage,
	}) => {
		// Login
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()

		await destinationsPage.navigateToDestinations()
		await destinationsPage.clickCreateDestination()

		// Test Amazon S3 connector
		await createDestinationPage.selectConnector("Amazon S3")
		await expect(
			createDestinationPage.page.locator("text=Amazon S3"),
		).toBeVisible()

		// Test Apache Iceberg connector
		await createDestinationPage.selectConnector("Apache Iceberg")
		await expect(
			createDestinationPage.page.locator("text=Apache Iceberg"),
		).toBeVisible()

		// Should show catalog selection for Iceberg
		await expect(createDestinationPage.catalogSelect).toBeVisible()
	})

	test("should support keyboard navigation throughout destination flow", async ({
		loginPage,
		destinationsPage,
		createDestinationPage,
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
		await destinationsPage.navigateToDestinations()
		await destinationsPage.clickCreateDestination()

		// Fill form using keyboard navigation
		await page.keyboard.press("Tab") // Navigate to destination name
		await page.keyboard.type("keyboard-test")

		// Can continue with Tab navigation through the form
		await expect(createDestinationPage.destinationNameInput).toHaveValue(
			"keyboard-test",
		)
	})
})
