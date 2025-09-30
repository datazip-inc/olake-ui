import { test, expect } from "../fixtures/auth.fixture"
import { MONGODB_TEST_CONFIG } from "../setup/test-env"

test.describe("Create Source Flow", () => {
	test.beforeEach(async ({ loginPage, sourcesPage }) => {
		// Login first
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()

		// Navigate to sources and create new source
		await sourcesPage.navigateToSources()
		await sourcesPage.clickCreateSource()
	})

	test("should display create source page correctly", async ({
		createSourcePage,
	}) => {
		await createSourcePage.expectCreateSourcePageVisible()
	})

	test("should create MongoDB source successfully", async ({
		createSourcePage,
		sourcesPage,
	}) => {
		const sourceData = {
			name: "testmongo",
			...MONGODB_TEST_CONFIG,
		}

		// Fill the MongoDB form
		await createSourcePage.fillMongoDBForm(sourceData)

		// Create the source
		await createSourcePage.clickCreate()

		// Wait for test connection and success modals
		await createSourcePage.expectTestConnectionModal()
		await createSourcePage.expectSuccessModal()
		await createSourcePage.expectEntitySavedModal()

		// Should navigate back to sources page
		await sourcesPage.expectSourcesPageVisible()
		await sourcesPage.expectSourceExists(sourceData.name)
	})

	test("should show validation error for empty source name", async ({
		createSourcePage,
	}) => {
		// Try to create without filling source name
		await createSourcePage.clickCreate()
		await createSourcePage.expectValidationError("Source name is required")
	})

	test("should show validation error for missing host", async ({
		createSourcePage,
	}) => {
		await createSourcePage.fillSourceName("test-source")
		await createSourcePage.clickCreate()

		// Should show validation for missing required fields
		await expect(
			createSourcePage.page.locator(".ant-form-item-has-error"),
		).toBeVisible()
	})

	test("should allow canceling source creation", async ({
		createSourcePage,
		sourcesPage,
	}) => {
		await createSourcePage.fillSourceName("test-cancel")
		await createSourcePage.clickCancel()

		// Should show confirmation modal and navigate back
		await expect(createSourcePage.page.getByText("Cancel")).toBeVisible()
		await createSourcePage.page.getByRole("button", { name: "Confirm" }).click()

		await sourcesPage.expectSourcesPageVisible()
	})

	test("should fill form step by step correctly", async ({
		createSourcePage,
	}) => {
		// Test step-by-step form filling to ensure all fields work
		await createSourcePage.fillSourceName("step-by-step-test")
		await expect(createSourcePage.sourceNameInput).toHaveValue(
			"step-by-step-test",
		)

		await createSourcePage.addHost("localhost:27017")
		await expect(createSourcePage.hostsInput).toHaveValue("")

		await createSourcePage.fillDatabaseName("testdb")
		await expect(createSourcePage.databaseInput).toHaveValue("testdb")

		await createSourcePage.fillCredentials("testuser", "testpass")
		await expect(createSourcePage.usernameInput).toHaveValue("testuser")
		await expect(createSourcePage.passwordInput).toHaveValue("testpass")
	})

	test("should handle connector selection", async ({ createSourcePage }) => {
		// Test connector dropdown functionality
		await createSourcePage.selectConnector("PostgreSQL")

		// The form should adapt to PostgreSQL schema
		await expect(createSourcePage.page.locator("text=PostgreSQL")).toBeVisible()
	})

	test("should support setup type selection", async ({ createSourcePage }) => {
		// Test new vs existing source setup
		await createSourcePage.selectSetupType("existing")
		await expect(createSourcePage.setupTypeExisting).toBeVisible()

		await createSourcePage.selectSetupType("new")
		await expect(createSourcePage.setupTypeNew).toBeVisible()
	})
})
