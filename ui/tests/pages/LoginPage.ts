import { Page, Locator, expect } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"

export class LoginPage {
	readonly page: Page
	readonly usernameInput: Locator
	readonly passwordInput: Locator
	readonly loginButton: Locator
	readonly errorMessage: Locator
	readonly pageTitle: Locator

	constructor(page: Page) {
		this.page = page
		this.usernameInput = page.getByPlaceholder("Username")
		this.passwordInput = page.getByPlaceholder("Password")
		this.loginButton = page.getByRole("button", { name: "Log in" })
		this.errorMessage = page.locator(".ant-message-error")
		this.pageTitle = page.locator("text=Login")
	}

	async goto() {
		await this.page.goto("/login")
	}

	async login(username: string, password: string) {
		await this.usernameInput.fill(username)
		await this.passwordInput.fill(password)
		await this.loginButton.click()
	}

	async expectLoginPageVisible() {
		await expect(this.pageTitle).toBeVisible()
		await expect(this.usernameInput).toBeVisible()
		await expect(this.passwordInput).toBeVisible()
		await expect(this.loginButton).toBeVisible()
	}

	async expectErrorMessage() {
		await expect(this.errorMessage).toBeVisible()
	}

	async waitForLogin() {
		await this.page.waitForURL("/jobs", { timeout: TIMEOUTS.SHORT })
	}

	async expectValidationError(message: string) {
		await expect(this.page.locator(`text=${message}`)).toBeVisible()
	}
}
