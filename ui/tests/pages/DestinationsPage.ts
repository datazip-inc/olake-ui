import { Page, Locator, expect } from "@playwright/test"
import { BasePage } from "./BasePage"

export class DestinationsPage extends BasePage {
	readonly createDestinationButton: Locator
	readonly destinationsTitle: Locator
	readonly destinationsLink: Locator
	readonly activeTab: Locator
	readonly inactiveTab: Locator
	readonly destinationTable: Locator

	constructor(page: Page) {
		super(page)
		this.createDestinationButton = page.getByRole("button", {
			name: "Create Destination",
		})
		this.destinationsTitle = page.locator("h1", { hasText: "Destinations" })
		this.destinationsLink = page.getByRole("link", { name: "Destinations" })
		this.activeTab = page.getByRole("tab", { name: "Active" })
		this.inactiveTab = page.getByRole("tab", { name: "Inactive" })
		this.destinationTable = page.locator(".ant-table-tbody")
	}

	async goto() {
		await super.goto("/destinations")
	}

	async navigateToDestinations() {
		await this.destinationsLink.click()
	}

	async clickCreateDestination() {
		await this.createDestinationButton.click()
	}

	async expectDestinationsPageVisible() {
		await this.expectVisible(this.destinationsTitle)
		await this.expectVisible(this.createDestinationButton)
	}

	async getDestinationRow(destinationName: string) {
		return this.page.getByRole("row", {
			name: new RegExp(destinationName, "i"),
		})
	}

	async editDestination(destinationName: string) {
		const destinationRow = await this.getDestinationRow(destinationName)
		await destinationRow.getByRole("button").click()
		await this.page.getByText("Edit").click()
	}

	async deleteDestination(destinationName: string) {
		const destinationRow = await this.getDestinationRow(destinationName)
		await destinationRow.getByRole("button").click()
		await this.page.getByText("Delete").click()
	}

	async expectDestinationExists(destinationName: string) {
		await this.switchToInactiveTab()
		const destinationRow = await this.getDestinationRow(destinationName)
		await expect(destinationRow).toBeVisible()
	}

	async expectDestinationNotExists(destinationName: string) {
		const destinationRow = await this.getDestinationRow(destinationName)
		await expect(destinationRow).not.toBeVisible()
	}

	async switchToInactiveTab() {
		await this.inactiveTab.click()
	}

	async switchToActiveTab() {
		await this.activeTab.click()
	}
}
