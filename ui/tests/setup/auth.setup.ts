import { test as setup } from "@playwright/test"
import { LoginPage } from "../pages/LoginPage"
import { LOGIN_CREDENTIALS } from "../utils"

const AUTH_STATE_PATH = "./tests/.auth/user.json"

/**
 * Setup authentication state
 * This runs once before all tests to establish authenticated state
 */
setup("authenticate", async ({ page }) => {
	const loginPage = new LoginPage(page)
	await loginPage.goto()
	await loginPage.login(
		LOGIN_CREDENTIALS.admin.username,
		LOGIN_CREDENTIALS.admin.password,
	)
	await loginPage.waitForLogin()

	await page.context().storageState({ path: AUTH_STATE_PATH })
})
