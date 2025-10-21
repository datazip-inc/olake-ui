import { Page, Locator, expect } from "@playwright/test"
import { BasePage } from "./BasePage"

export class EditSourcePage extends BasePage {
	readonly sourceNameInput: Locator
	readonly saveChangesButton: Locator
	readonly confirmButton: Locator
	readonly cancelButton: Locator
	readonly associatedJobsButton: Locator
	readonly backToSourcesButton: Locator
	readonly pageTitle: Locator
	readonly deleteButton: Locator
	readonly testConnectionButton: Locator

	constructor(page: Page) {
		super(page)
		this.sourceNameInput = page.getByPlaceholder(
			"Enter the name of your source",
		)
		this.saveChangesButton = page.getByRole("button", { name: "Save changes" })
		this.confirmButton = page.getByRole("button", { name: "Confirm" })
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.associatedJobsButton = page.getByRole("button", {
			name: "Associated jobs",
		})
		this.backToSourcesButton = page.getByRole("button", { name: "Sources" })
		this.pageTitle = page.locator("text=Edit source")
		this.deleteButton = page.getByRole("button", { name: "Delete" })
		this.testConnectionButton = page.getByRole("button", {
			name: "Test Connection",
		})
	}

	async goto(sourceId: string) {
		await super.goto(`/sources/${sourceId}`)
	}

	async expectEditSourcePageVisible() {
		await this.expectVisible(this.pageTitle)
		await this.expectVisible(this.saveChangesButton)
	}

	async updateSourceName(newName: string) {
		await this.sourceNameInput.clear()
		await this.sourceNameInput.fill(newName)
	}

	async clickSaveChanges() {
		await this.saveChangesButton.click()
	}

	async clickConfirm() {
		await this.confirmButton.click()
	}

	async clickCancel() {
		await this.cancelButton.click()
	}

	async viewAssociatedJobs() {
		await this.associatedJobsButton.click()
	}

	async goBackToSources() {
		await this.backToSourcesButton.click()
	}

	async deleteSource() {
		await this.deleteButton.click()
	}

	async testConnection() {
		await this.testConnectionButton.click()
	}

	async expectSuccessMessage() {
		await expect(this.page.locator(".ant-message-success")).toBeVisible()
	}

	async expectValidationError(message: string) {
		await expect(this.page.locator(`text=${message}`)).toBeVisible()
	}

	async expectConfirmationDialog() {
		await expect(this.confirmButton).toBeVisible()
	}
}
