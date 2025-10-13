import { Page, Locator, expect } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"
import { BasePage } from "./BasePage"

export class JobsPage extends BasePage {
	readonly createJobButton: Locator
	readonly jobsTitle: Locator
	readonly jobsLink: Locator
	readonly activeTab: Locator
	readonly inactiveTab: Locator
	readonly savedTab: Locator
	readonly jobTable: Locator

	constructor(page: Page) {
		super(page)
		this.createJobButton = page.getByRole("button", { name: "Create Job" })
		this.jobsTitle = page.locator("h1", { hasText: "Jobs" })
		this.jobsLink = page.getByRole("link", { name: "Jobs" })
		this.activeTab = page.getByRole("tab", { name: "Active" })
		this.inactiveTab = page.getByRole("tab", { name: "Inactive" })
		this.savedTab = page.getByRole("tab", { name: "Saved" })
		this.jobTable = page.locator(".ant-table-tbody")
	}

	async goto() {
		await super.goto("/jobs")
	}

	async navigateToJobs() {
		await this.jobsLink.click()
	}

	async clickCreateJob() {
		await this.createJobButton.click()
	}

	async expectJobsPageVisible() {
		await this.expectVisible(this.jobsTitle)
		await this.expectVisible(this.createJobButton)
	}

	async getJobRow(jobName: string) {
		return this.page.getByTestId(`job-${jobName}`)
	}

	async syncJob(jobName: string) {
		const jobRow = await this.getJobRow(jobName)
		await jobRow.click()
		await this.page.getByText("Sync now").click()
	}

	async editJob(jobName: string) {
		const jobRow = await this.getJobRow(jobName)
		await jobRow.getByRole("button").click()
		await this.page.getByText("Edit").click()
	}

	async viewLogs(jobName: string) {
		const jobRow = await this.getJobRow(jobName)
		await jobRow.getByRole("button").click()
		await this.page.getByText("View logs").click()
	}

	async expectJobExists(jobName: string) {
		const jobRow = this.page.getByTestId(`job-${jobName}`)
		await expect(jobRow).toBeVisible()
	}

	async expectJobNotExists(jobName: string) {
		const jobRow = await this.getJobRow(jobName)
		await expect(jobRow).not.toBeVisible()
	}

	async switchToInactiveTab() {
		await this.inactiveTab.click()
	}

	async switchToActiveTab() {
		await this.activeTab.click()
	}

	async switchToSavedTab() {
		await this.savedTab.click()
	}

	async viewJobLogs() {
		// Wait for the page to be fully loaded first
		await this.page.waitForLoadState("networkidle", {
			timeout: TIMEOUTS.LONG,
		})

		// Wait for the View logs button to be visible and enabled
		const viewLogsButton = this.page.getByRole("button", { name: "View logs" })
		await viewLogsButton.waitFor({
			state: "visible",
			timeout: TIMEOUTS.LONG,
		})

		// Ensure the button is enabled before clicking
		await expect(viewLogsButton).toBeEnabled({ timeout: TIMEOUTS.MEDIUM })

		await viewLogsButton.click()
	}

	async viewJobConfigurations() {
		await this.page
			.getByRole("button", { name: "View job configurations" })
			.click()
	}

	async expectLogsCellVisible() {
		await expect(
			this.page.getByRole("cell", { name: "Total records read:" }),
		).toBeVisible()
	}
}
