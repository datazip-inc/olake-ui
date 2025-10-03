import { test, expect } from "../fixtures/auth.fixture"
import { TEST_CREDENTIALS, VALIDATION_MESSAGES, URLS } from "../setup/test-env"

test.describe("Login Authentication", () => {
	test.beforeEach(async ({ loginPage }) => {
		await loginPage.goto()
	})

	test("should display login page correctly", async ({ loginPage }) => {
		await loginPage.expectLoginPageVisible()
	})

	test("should login successfully with admin credentials", async ({
		loginPage,
		page,
	}) => {
		await loginPage.login(
			TEST_CREDENTIALS.admin.username,
			TEST_CREDENTIALS.admin.password,
		)
		await loginPage.waitForLogin()
		await expect(page).toHaveURL(URLS.jobs)
		await expect(page.locator("h1, h2, h3").first()).toBeVisible()
	})

	test("should show error for invalid credentials", async ({ loginPage }) => {
		await loginPage.login(
			TEST_CREDENTIALS.invalid.username,
			TEST_CREDENTIALS.invalid.password,
		)
		await loginPage.expectErrorMessage()
	})

	test("should show validation errors for empty fields", async ({
		loginPage,
	}) => {
		await loginPage.loginButton.click()
		await loginPage.expectValidationError(VALIDATION_MESSAGES.username.required)

		await loginPage.usernameInput.fill(TEST_CREDENTIALS.admin.username)
		await loginPage.loginButton.click()
		await loginPage.expectValidationError(VALIDATION_MESSAGES.password.required)
	})

	test("should show validation errors for short inputs", async ({
		loginPage,
	}) => {
		await loginPage.usernameInput.fill("ab")
		await loginPage.passwordInput.fill("12345")
		await loginPage.loginButton.click()

		await loginPage.expectValidationError(
			VALIDATION_MESSAGES.username.minLength,
		)
		await loginPage.expectValidationError(
			VALIDATION_MESSAGES.password.minLength,
		)
	})

	test("should clear form after failed login attempt", async ({
		loginPage,
	}) => {
		await loginPage.login(
			TEST_CREDENTIALS.invalid.username,
			TEST_CREDENTIALS.invalid.password,
		)
		await loginPage.expectErrorMessage()

		await expect(loginPage.usernameInput).toHaveValue("")
		await expect(loginPage.passwordInput).toHaveValue("")
	})

	test("should handle keyboard navigation", async ({ loginPage, page }) => {
		await page.keyboard.press("Tab")
		await expect(loginPage.usernameInput).toBeFocused()

		await page.keyboard.press("Tab")
		await expect(loginPage.passwordInput).toBeFocused()

		await page.keyboard.press("Tab")
		await expect(loginPage.loginButton).toBeFocused()
	})

	test("should submit form with Enter key", async ({ loginPage, page }) => {
		await loginPage.usernameInput.fill(TEST_CREDENTIALS.admin.username)
		await loginPage.passwordInput.fill(TEST_CREDENTIALS.admin.password)
		await loginPage.passwordInput.press("Enter")

		await loginPage.waitForLogin()
		await expect(page).toHaveURL(URLS.jobs)
	})
})
