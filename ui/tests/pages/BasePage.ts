import { Page, Locator, expect } from "@playwright/test"

/**
 * Base Page Object class containing common methods and utilities
 * that are shared across all page objects.
 *
 * All page objects should extend this class to inherit common functionality
 * and maintain consistency across the test suite.
 */
export abstract class BasePage {
	readonly page: Page

	constructor(page: Page) {
		this.page = page
	}

	/**
	 * Navigate to a specific URL
	 * @param url - The URL path to navigate to (relative to baseURL)
	 */
	async goto(url: string): Promise<void> {
		await this.page.goto(url)
	}

	/**
	 * Wait for the page to fully load
	 * @param state - The load state to wait for (default: "networkidle")
	 */
	async waitForPageLoad(
		state: "load" | "domcontentloaded" | "networkidle" = "networkidle",
	): Promise<void> {
		await this.page.waitForLoadState(state)
	}

	/**
	 * Click a button by its accessible name
	 * @param name - The accessible name of the button
	 */
	async clickButton(name: string): Promise<void> {
		await this.page.getByRole("button", { name }).click()
	}

	/**
	 * Click a link by its accessible name
	 * @param name - The accessible name of the link
	 */
	async clickLink(name: string): Promise<void> {
		await this.page.getByRole("link", { name }).click()
	}

	/**
	 * Get an element by its test ID
	 * @param testId - The test ID attribute value
	 */
	getByTestId(testId: string): Locator {
		return this.page.getByTestId(testId)
	}

	/**
	 * Get an element by its role
	 * @param role - The ARIA role
	 * @param options - Additional options like name, exact, etc.
	 */
	getByRole(
		role:
			| "button"
			| "link"
			| "textbox"
			| "checkbox"
			| "radio"
			| "tab"
			| "row"
			| "cell"
			| "heading"
			| "list"
			| "listitem"
			| "table"
			| "switch",
		options?: { name?: string | RegExp; exact?: boolean },
	): Locator {
		return this.page.getByRole(role, options)
	}

	/**
	 * Expect a validation error message to be visible
	 * @param message - The error message text
	 */
	async expectValidationError(message: string): Promise<void> {
		await expect(this.page.locator(`text=${message}`)).toBeVisible()
	}

	/**
	 * Expect an element to be visible
	 * @param locator - The locator to check
	 */
	async expectVisible(locator: Locator): Promise<void> {
		await expect(locator).toBeVisible()
	}

	/**
	 * Expect an element to be hidden
	 * @param locator - The locator to check
	 */
	async expectHidden(locator: Locator): Promise<void> {
		await expect(locator).not.toBeVisible()
	}

	/**
	 * Fill an input field
	 * @param locator - The input field locator
	 * @param value - The value to fill
	 */
	async fillInput(locator: Locator, value: string): Promise<void> {
		await locator.click()
		await locator.fill(value)
	}

	/**
	 * Wait for URL to match a pattern
	 * @param pattern - The URL pattern (string or regex)
	 * @param options - Wait options
	 */
	async waitForURL(
		pattern: string | RegExp,
		options?: { timeout?: number },
	): Promise<void> {
		await this.page.waitForURL(pattern, options)
	}

	/**
	 * Wait for a selector to be visible
	 * @param selector - The CSS selector or text selector
	 * @param options - Wait options
	 */
	async waitForSelector(
		selector: string,
		options?: { state?: "visible" | "hidden"; timeout?: number },
	): Promise<void> {
		await this.page.waitForSelector(selector, {
			state: options?.state || "visible",
			timeout: options?.timeout,
		})
	}

	/**
	 * Get text content from an element
	 * @param locator - The element locator
	 */
	async getTextContent(locator: Locator): Promise<string | null> {
		return await locator.textContent()
	}

	/**
	 * Check if an element is visible
	 * @param locator - The element locator
	 */
	async isVisible(locator: Locator): Promise<boolean> {
		return await locator.isVisible()
	}

	/**
	 * Check if an element is enabled
	 * @param locator - The element locator
	 */
	async isEnabled(locator: Locator): Promise<boolean> {
		return await locator.isEnabled()
	}
}
