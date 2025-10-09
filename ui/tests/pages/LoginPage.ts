import { Page, Locator } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"
import { BasePage } from "./BasePage"

export class LoginPage extends BasePage {
	readonly usernameInput: Locator
	readonly passwordInput: Locator
	readonly loginButton: Locator
	readonly errorMessage: Locator
	readonly pageTitle: Locator

	constructor(page: Page) {
		super(page)
		this.usernameInput = page.getByPlaceholder("Username")
		this.passwordInput = page.getByPlaceholder("Password")
		this.loginButton = page.getByRole("button", { name: "Log in" })
		this.errorMessage = page.locator(".ant-message-error")
		this.pageTitle = page.locator("text=Login")
	}

	async goto() {
		await super.goto("/login")
	}

	async login(username: string, password: string) {
		await this.usernameInput.fill(username)
		await this.passwordInput.fill(password)
		await this.loginButton.click()
	}

	async expectLoginPageVisible() {
		await this.expectVisible(this.pageTitle)
		await this.expectVisible(this.usernameInput)
		await this.expectVisible(this.passwordInput)
		await this.expectVisible(this.loginButton)
	}

	async expectErrorMessage() {
		await this.expectVisible(this.errorMessage)
	}

	async waitForLogin() {
		await this.waitForURL("/jobs", { timeout: TIMEOUTS.SHORT })
	}
}
