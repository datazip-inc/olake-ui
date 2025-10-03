import { Page, Locator, expect } from "@playwright/test"

export class SourcesPage {
	readonly page: Page
	readonly createSourceButton: Locator
	readonly sourcesTitle: Locator
	readonly sourcesLink: Locator
	readonly activeTab: Locator
	readonly inactiveTab: Locator
	readonly sourceTable: Locator

	constructor(page: Page) {
		this.page = page
		this.createSourceButton = page.getByRole("button", {
			name: "Create Source",
		})
		this.sourcesTitle = page.locator("h1", { hasText: "Sources" })
		this.sourcesLink = page.getByRole("link", { name: "Sources" })
		this.activeTab = page.getByRole("tab", { name: "Active" })
		this.inactiveTab = page.getByRole("tab", { name: "Inactive" })
		this.sourceTable = page.locator(".ant-table-tbody")
	}

	async goto() {
		await this.page.goto("/sources")
	}

	async navigateToSources() {
		await this.sourcesLink.click()
	}

	async clickCreateSource() {
		await this.createSourceButton.click()
	}

	async expectSourcesPageVisible() {
		await expect(this.sourcesTitle).toBeVisible()
		await expect(this.createSourceButton).toBeVisible()
	}

	async getSourceRow(sourceName: string) {
		return this.page.getByRole("row", { name: new RegExp(sourceName, "i") })
	}

	async editSource(sourceName: string) {
		const sourceRow = await this.getSourceRow(sourceName)
		await sourceRow.getByRole("button").click()
		await this.page.getByText("Edit").click()
	}

	async deleteSource(sourceName: string) {
		const sourceRow = await this.getSourceRow(sourceName)
		await sourceRow.getByRole("button").click()
		await this.page.getByText("Delete").click()
	}

	async expectSourceExists(sourceName: string) {
		const sourceRow = await this.getSourceRow(sourceName)
		await expect(sourceRow).toBeVisible()
	}

	async expectSourceNotExists(sourceName: string) {
		const sourceRow = await this.getSourceRow(sourceName)
		await expect(sourceRow).not.toBeVisible()
	}

	async switchToInactiveTab() {
		await this.inactiveTab.click()
	}

	async switchToActiveTab() {
		await this.activeTab.click()
	}
}
