import { Page, Locator, expect } from "@playwright/test"
import { BasePage } from "./BasePage"

export class EditDestinationPage extends BasePage {
	readonly destinationNameInput: Locator
	readonly saveChangesButton: Locator
	readonly confirmButton: Locator
	readonly cancelButton: Locator
	readonly associatedJobsButton: Locator
	readonly configButton: Locator
	readonly backToDestinationsButton: Locator
	readonly pageTitle: Locator
	readonly deleteButton: Locator
	readonly testConnectionButton: Locator

	constructor(page: Page) {
		super(page)
		this.destinationNameInput = page.getByPlaceholder(
			"Enter the name of your destination",
		)
		this.saveChangesButton = page.getByRole("button", { name: "Save Changes" })
		this.confirmButton = page.getByRole("button", { name: "Confirm" })
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.associatedJobsButton = page.getByRole("button", {
			name: "Associated jobs",
		})
		this.configButton = page.getByRole("button", { name: "Config" })
		this.backToDestinationsButton = page.getByRole("button", {
			name: "Destinations",
		})
		this.pageTitle = page.locator("text=Edit destination")
		this.deleteButton = page.getByRole("button", { name: "Delete" })
		this.testConnectionButton = page.getByRole("button", {
			name: "Test Connection",
		})
	}

	async goto(destinationId: string) {
		await super.goto(`/destinations/${destinationId}`)
	}

	async expectEditDestinationPageVisible() {
		await this.expectVisible(this.pageTitle)
		await this.expectVisible(this.saveChangesButton)
	}

	async updateDestinationName(newName: string) {
		await this.destinationNameInput.clear()
		await this.destinationNameInput.fill(newName)
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

	async viewConfig() {
		await this.configButton.click()
	}

	async goBackToDestinations() {
		await this.backToDestinationsButton.click()
	}

	async deleteDestination() {
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
