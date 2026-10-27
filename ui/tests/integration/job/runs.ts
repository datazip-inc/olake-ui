/** A job's runs: the jobs page actions, Sync now, and the history page. */

import { Page } from "@playwright/test"

import { expect } from "../../fixtures"
import { JobsPage } from "../../pages/JobsPage"
import { SYNC_TIMEOUT, TIMEOUTS } from "../e2e"

export type RunType = "Sync" | "Clear Destination"
export type RunStatus = "Completed" | "Failed"

/** Opens one entry of a job's actions menu on the jobs page. */
export const jobMenu = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
	item: string,
) => {
	await jobsPage.navigateToJobs()
	await jobsPage.expectJobsPageVisible()
	await page.getByTestId(`job-${job}`).click()
	await page.getByRole("menuitem", { name: item }).click()
}

/** The rows of the history table. */
export const historyRows = (page: Page) =>
	page.locator(".ant-table-tbody tr.ant-table-row")

/**
 * On the job's history page, waits until it lists runs rows and the newest has jobType and
 * status. Refreshes the table between tries.
 */
export const waitForNewestRun = async (
	page: Page,
	runs: number,
	jobType: RunType,
	status: RunStatus,
	timeout = SYNC_TIMEOUT,
) => {
	await expect(async () => {
		await page.getByTestId("refresh-tasks-button").click()
		await expect(historyRows(page)).toHaveCount(runs, { timeout: 5000 })
		const newest = historyRows(page).first()
		await expect(newest).toContainText(jobType, { timeout: 2000 })
		await expect(newest).toContainText(status, { timeout: 2000 })
	}).toPass({ timeout, intervals: [3000] })
}

/** "Sync now" from the jobs page, then waits for the run on the history page. */
export const syncAndWait = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
	runs: number,
	status: RunStatus,
) => {
	await jobsPage.navigateToJobs()
	await jobsPage.expectJobsPageVisible()
	await jobsPage.syncJob(job)
	await waitForNewestRun(page, runs, "Sync", status)
}

/**
 * On the history page, checks the table: its columns, that it lists runs (newest first) all
 * Completed, the search by status, and the refresh button.
 */
export const expectHistory = async (page: Page, runs: RunType[]) => {
	for (const header of [
		"Start time",
		"Runtime",
		"Status",
		"Job Type",
		"Actions",
	]) {
		await expect(page.getByRole("columnheader", { name: header })).toBeVisible()
	}
	const rows = historyRows(page)
	await expect(rows).toHaveCount(runs.length)
	await expect(rows.filter({ hasText: "Completed" })).toHaveCount(runs.length)
	await expect(rows.filter({ hasText: "Clear Destination" })).toHaveCount(
		runs.filter(run => run === "Clear Destination").length,
	)
	for (const [index, run] of runs.entries()) {
		await expect(rows.nth(index)).toContainText(run)
	}

	// Search matches the status and the start time.
	const search = page.getByPlaceholder("Search Tasks")
	await search.fill("completed")
	await expect(rows).toHaveCount(runs.length)
	await search.fill("no-such-run")
	await expect(rows).toHaveCount(0)
	await search.fill("")
	await expect(rows).toHaveCount(runs.length)

	await page.getByTestId("refresh-tasks-button").click()
	await expect(rows).toHaveCount(runs.length)
}

/** On the jobs page, the "Failed jobs" tab lists job with its failed sync. */
export const expectInFailedJobs = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
) => {
	await jobsPage.navigateToJobs()
	await jobsPage.expectJobsPageVisible()
	await page.getByRole("tab", { name: "Failed jobs" }).click()
	// antd keeps the panels it has shown in the page, so look only in the Failed jobs panel.
	const failed = page.getByRole("tabpanel", { name: "Failed jobs" })
	const row = failed.getByRole("row").filter({ hasText: job })
	await expect(row).toBeVisible({ timeout: TIMEOUTS.LONG })
	await expect(row).toContainText("Failed")
	await expect(row).toContainText("Sync")
	await expect(failed.getByTestId(`job-${job}`)).toBeVisible()
}
