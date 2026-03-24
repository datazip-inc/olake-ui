import { Page, Locator, expect } from "@playwright/test"

import { BasePage } from "./BasePage"

export class SourcesPage extends BasePage {
	readonly createSourceButton: Locator
	readonly sourcesTitle: Locator
	readonly sourcesLink: Locator
	readonly activeTab: Locator
	readonly inactiveTab: Locator
	readonly sourceTable: Locator

	constructor(page: Page) {
		super(page)
		this.createSourceButton = page.getByTestId("create-source-button")
		this.sourcesTitle = page.locator("h1", { hasText: "Sources" })
		this.sourcesLink = page.getByRole("link", { name: "Sources" })
		this.activeTab = page.getByRole("tab", { name: "Active" })
		this.inactiveTab = page.getByRole("tab", { name: "Inactive" })
		this.sourceTable = page.locator(".ant-table-tbody")
	}

	async goto() {
		await super.goto("/sources")
	}

	async navigateToSources() {
		await this.sourcesLink.click()
	}

	async clickCreateSource() {
		// Wait until the sources query has finished — the modal only mounts after
		// loading completes, so any earlier check is a race condition.
		await this.page
			.locator('[data-testid="sources-page"][data-loaded="true"]')
			.waitFor({ timeout: 60_000 })

		const onboardingCta = this.page.getByTestId(
			"onboarding-create-source-button",
		)
		if (await onboardingCta.isVisible()) {
			await onboardingCta.click()
		} else {
			await this.createSourceButton.click()
		}
		await this.page.waitForURL(/\/sources\/new/)
	}

	async expectSourcesPageVisible() {
		await this.expectVisible(this.sourcesTitle)
		await this.expectVisible(this.createSourceButton)
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
		await this.switchToInactiveTab()
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
