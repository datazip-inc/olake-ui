import { Page, Locator, expect } from "@playwright/test"

import { BasePage } from "./BasePage"

export class SettingsPage extends BasePage {
	readonly webhookInput: Locator
	readonly saveButton: Locator
	readonly editButton: Locator
	readonly clearButton: Locator

	constructor(page: Page) {
		super(page)
		this.webhookInput = page.getByPlaceholder("Enter your webhook URL")
		this.saveButton = page.getByRole("button", { name: "Save changes" })
		this.editButton = page.getByRole("button", { name: "Edit" })
		this.clearButton = page.getByRole("button", { name: "Clear webhook URL" })
	}

	async goto() {
		await super.goto("/settings")
		await expect(
			this.page.getByText("Webhook Alerts", { exact: true }),
		).toBeVisible()
	}

	// Fills the URL and saves it; the field is read-only once saved.
	async saveWebhook(url: string) {
		// a saved URL is read-only until Edit is clicked
		if (await this.editButton.isVisible()) await this.editButton.click()
		await this.webhookInput.fill(url)
		await this.saveButton.click()
		await expect(this.editButton).toBeVisible()
		await expect(this.webhookInput).toHaveValue(url)
		await expect(this.webhookInput).toBeDisabled()
	}

	// Does nothing when no URL is set, as the Clear button only exists for a saved URL.
	async clearWebhook() {
		if (await this.clearButton.isVisible()) await this.clearButton.click()
		await expect(this.webhookInput).toHaveValue("")
		await expect(this.webhookInput).toBeEnabled()
	}
}
