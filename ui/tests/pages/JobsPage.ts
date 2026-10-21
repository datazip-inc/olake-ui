import { Page, Locator, expect } from "@playwright/test"

import { BasePage } from "./BasePage"
import { TIMEOUTS } from "../../playwright.config"
import { pollToClickAndVerifyText } from "../utils/page-utils"

export class JobsPage extends BasePage {
	readonly createJobButton: Locator
	readonly jobsTitle: Locator
	readonly jobsLink: Locator
	readonly activeTab: Locator
	readonly inactiveTab: Locator
	readonly savedTab: Locator
	readonly failedTab: Locator
	readonly jobTable: Locator

	constructor(page: Page) {
		super(page)
		this.createJobButton = page.getByTestId("create-job-button")
		this.jobsTitle = page.locator("h1", { hasText: "Jobs" })
		this.jobsLink = page.getByRole("link", { name: "Jobs" })
		this.activeTab = page.getByRole("tab", { name: "Active jobs", exact: true })
		this.inactiveTab = page.getByRole("tab", {
			name: "Inactive jobs",
			exact: true,
		})
		this.savedTab = page.getByRole("tab", { name: "Saved jobs", exact: true })
		this.failedTab = page.getByRole("tab", { name: "Failed jobs" })
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
		const syncNowBtn = this.page.getByText("Sync now")
		await expect(syncNowBtn).toBeVisible()
		await expect(syncNowBtn).toBeEnabled()

		// Click the button AND wait for the navigation to the history page
		await Promise.all([
			this.page.waitForURL("**/history", { timeout: 30000 }), // Wait up to 30s for the redirect, which is slow while an edit clears its new streams
			syncNowBtn.click(),
		])
	}

	// Syncs the job from the jobs list and waits until its task logs report completion.
	async syncAndWait(jobName: string) {
		await this.goto()
		await this.syncJob(jobName)
		await this.viewJobLogs()
		await this.waitForSyncCompletionLogs()
		await this.navigateToJobs()
		await this.expectJobStatus(jobName, "Completed")
	}

	// Opens the job's streams in edit mode, from its actions menu.
	async editStreams(jobName: string) {
		const jobRow = await this.getJobRow(jobName)
		await jobRow.click()
		await this.page.getByText("Edit Streams").click()
		await expect(this.page.getByText("Sync all")).toBeVisible()
	}

	async openHistory(jobName: string) {
		const jobRow = await this.getJobRow(jobName)
		await jobRow.click()
		await this.page.getByText("Job Logs & History").click()
		await this.page.waitForURL("**/history")
	}

	// The job's actions menu trigger inside one tab's panel; a tab keeps its list from when it was last
	// opened, so the same job can still be listed in a tab it has left.
	private tabRow(tab: "Active jobs" | "Inactive jobs", jobName: string) {
		return this.page.getByLabel(tab).getByTestId(`job-${jobName}`).first()
	}

	// Pauses the job from its actions menu; it moves to the Inactive jobs tab.
	async pauseJob(jobName: string) {
		await this.tabRow("Active jobs", jobName).click()
		await this.page.getByText("Pause job").click()
	}

	// Resumes a paused job from the Inactive jobs tab.
	async resumeJob(jobName: string) {
		await this.switchToInactiveTab()
		await this.tabRow("Inactive jobs", jobName).click()
		await this.page.getByText("Resume job").click()
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

	async switchToFailedTab() {
		await this.failedTab.click()
	}

	// Every tab keeps its table mounted, so look only inside the panel of the tab.
	async expectJobInTab(
		tab: "Active jobs" | "Inactive jobs" | "Failed jobs",
		jobName: string,
	) {
		await expect(
			this.page.getByLabel(tab).getByTestId(`job-${jobName}`).first(),
		).toBeVisible()
	}

	async expectJobNotInTab(
		tab: "Active jobs" | "Inactive jobs" | "Failed jobs",
		jobName: string,
	) {
		await expect(
			this.page.getByLabel(tab).getByTestId(`job-${jobName}`),
		).toHaveCount(0)
	}

	// Clears everything the job synced to the destination, from the job's settings page.
	async clearDestination(jobName: string) {
		// the row's test id sits on its actions menu trigger
		const jobRow = await this.getJobRow(jobName)
		await jobRow.click()
		await this.page.getByText("Job settings").click()
		// the server refuses while the last sync is still winding down, so ask again until it accepts
		await expect(async () => {
			await this.page.getByRole("button", { name: "Clear Destination" }).click()
			await this.page
				.getByRole("dialog")
				.getByRole("button", { name: "Confirm" })
				.click()
			await this.page.waitForURL("**/history", { timeout: 5000 })
		}).toPass({ timeout: TIMEOUTS.SHORT * 6, intervals: [5000] })
	}

	// Waits on the job's history page until a run of the type reaches the status.
	async waitForRun(runType: "Sync" | "Clear Destination", status: string) {
		await pollToClickAndVerifyText(
			this.page,
			this.page.getByTestId("refresh-tasks-button"),
			this.page
				.getByRole("row")
				.filter({ hasText: runType })
				.filter({ hasText: status })
				.first(),
		)
	}

	async switchToSavedTab() {
		await this.savedTab.click()
	}

	async viewJobLogs(jobName?: string) {
		await this.page.waitForLoadState("networkidle", {
			timeout: TIMEOUTS.LONG,
		})

		// Dynamic getter so Playwright re-evaluates the DOM on every retry loop
		const getViewLogsButton = () => {
			const container = jobName
				? this.page.getByTestId(`job-${jobName}`)
				: this.page
			return container.getByRole("button", { name: "View logs" }).first()
		}

		// Playwright will retry this block until the assertions inside pass
		await expect(async () => {
			const btn = getViewLogsButton()

			// If the job hasn't appeared yet, force a hard reload
			if (!(await btn.isVisible())) {
				await this.page.reload()
				await this.page.waitForLoadState("networkidle")
			}

			// Attempt to assert visibility. If this fails, expect.toPass() catches the error and retries.
			await expect(btn).toBeVisible({ timeout: 2000 })
		}).toPass({
			timeout: TIMEOUTS.LONG,
			intervals: [3000], // Wait 3 seconds between reload attempts
		})

		const finalBtn = getViewLogsButton()
		await expect(finalBtn).toBeEnabled({ timeout: TIMEOUTS.LONG })
		await finalBtn.click()
	}

	async viewJobConfigurations() {
		await this.page
			.getByRole("button", { name: "View job configurations" })
			.click()
	}

	// The line a sync writes when it has finished counting what it moved.
	get logSummary(): Locator {
		return this.page.getByText(/Total records read: \d+/)
	}

	async expectLogSummary() {
		await expect(this.logSummary.first()).toBeVisible()
	}

	async searchLogs(text: string) {
		await this.page.getByPlaceholder("Search Logs").fill(text)
	}

	async refreshLogs() {
		await this.page.getByTestId("refresh-logs-button").click()
	}

	async toggleErrorsOnly() {
		await this.page.getByRole("button", { name: "Errors", exact: true }).click()
	}

	async waitForSyncCompletionLogs() {
		// Find the refresh button using test id
		const refreshButton = this.page.getByTestId("refresh-logs-button")

		// Poll every 5 seconds for up to the LONG timeout
		await pollToClickAndVerifyText(
			this.page,
			refreshButton,
			"Sync completed, wait 5 seconds cleanup in progress...",
		)

		// Wait 5 seconds after the message appears as requested
		await this.page.waitForTimeout(5000)
	}

	async expectJobStatus(jobName: string, expectedStatus: string) {
		const row = this.page.getByRole("row").filter({ hasText: jobName }).first()
		await expect(row).toContainText(expectedStatus)
	}
}
