import { test as baseTest } from "./base.fixture"
import { LOGIN_CREDENTIALS } from "../utils"

/**
 * Authenticated test - automatically logs in before each test
 */
export const test = baseTest

// Auto-login before each test
test.beforeEach(async ({ loginPage }) => {
	await loginPage.goto()
	await loginPage.login(
		LOGIN_CREDENTIALS.admin.username,
		LOGIN_CREDENTIALS.admin.password,
	)
	await loginPage.waitForLogin()
})

export { expect } from "@playwright/test"
