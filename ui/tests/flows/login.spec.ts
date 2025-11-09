import { test, expect } from "../fixtures/auth.fixture"

test.describe("Login Flow", () => {
	test.beforeEach(async ({ loginPage }) => {
		await loginPage.goto()
	})

	test("should display login page correctly", async ({ loginPage }) => {
		await loginPage.expectLoginPageVisible()
	})

	test("should login successfully with valid credentials", async ({
		loginPage,
		page,
	}) => {
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()
		await expect(page).toHaveURL("/jobs")
	})

	test("should show error for invalid credentials", async ({ loginPage }) => {
		await loginPage.login("invalid", "invalid")
		await loginPage.expectErrorMessage()
	})

	test("should show validation errors for empty fields", async ({
		loginPage,
	}) => {
		await loginPage.loginButton.click()
		await loginPage.expectValidationError("Please input your username!")

		await loginPage.usernameInput.fill("admin")
		await loginPage.loginButton.click()
		await loginPage.expectValidationError("Please input your password!")
	})

	test("should show validation errors for short inputs", async ({
		loginPage,
	}) => {
		await loginPage.usernameInput.fill("ab")
		await loginPage.passwordInput.fill("12345")
		await loginPage.loginButton.click()

		await loginPage.expectValidationError(
			"Username must be at least 3 characters",
		)
		await loginPage.expectValidationError(
			"Password must be at least 6 characters",
		)
	})
})
