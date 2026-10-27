/** A job's runs: the jobs page actions, Sync now, and the history page. */

import { Page } from "@playwright/test"

import { expect } from "../../fixtures"
import { JobsPage } from "../../pages/JobsPage"
import { SYNC_TIMEOUT, TIMEOUTS } from "../e2e"

export type RunType = "Sync" | "Clear Destination"
export type RunStatus = "Completed" | "Failed"

/** A run as the history table shows it. */
export type HistoryRun = { type: RunType; status: RunStatus }

/** The history table shows this many runs a page, newest first. */
export const HISTORY_PAGE_SIZE = 8

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

/** The rows of the history table's current page. */
export const historyRows = (page: Page) =>
	page.locator(".ant-table-tbody tr.ant-table-row")

/**
 * On the job's history page, waits until the job has runs runs and the newest has jobType and
 * status. Refreshes the table between tries, and fails at once when the newest run of jobType
 * ended with another status. Past a page of runs only the page's count shows; after Sync now the
 * page itself waits for the new run before it lists any.
 */
export const waitForNewestRun = async (
	page: Page,
	runs: number,
	jobType: RunType,
	status: RunStatus,
	timeout = SYNC_TIMEOUT,
) => {
	const rows = historyRows(page)
	const deadline = Date.now() + timeout
	let newest = "(no runs listed)"
	for (;;) {
		await page.getByTestId("refresh-tasks-button").click()
		await page.waitForTimeout(1000)
		if ((await rows.count()) === Math.min(runs, HISTORY_PAGE_SIZE)) {
			newest = (await rows.first().innerText()).replace(/\s+/g, " ")
			if (newest.includes(jobType)) {
				if (newest.includes(status)) return
				const ended = ENDED.find(
					other => other !== status && newest.includes(other),
				)
				if (ended) {
					throw new Error(
						`the newest ${jobType} run ended ${ended}, want ${status}: ${newest}`,
					)
				}
			}
		}
		if (Date.now() > deadline) {
			throw new Error(
				`no ${status} ${jobType} run as the newest of ${runs} within ${timeout / 1000}s; newest: ${newest}`,
			)
		}
		await page.waitForTimeout(2000)
	}
}

/** The statuses of a run that has ended. */
const ENDED = ["Completed", "Failed", "Cancelled"]

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
 * On the history page, checks the table: its columns; its pages, which list runs (newest first)
 * HISTORY_PAGE_SIZE at a time, each with its Job Type and status; the search, which matches the
 * status; and the refresh button.
 */
export const expectHistory = async (page: Page, runs: HistoryRun[]) => {
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
	const expectRows = async (expected: HistoryRun[]) => {
		await expect(rows).toHaveCount(expected.length)
		for (const [index, run] of expected.entries()) {
			await expect(rows.nth(index)).toContainText(run.type)
			await expect(rows.nth(index)).toContainText(run.status)
		}
	}
	const pageButton = (n: number) => page.locator(`.ant-pagination-item-${n}`)

	const pages = Math.ceil(runs.length / HISTORY_PAGE_SIZE)
	for (let n = 1; n <= pages; n++) {
		if (n > 1) {
			await pageButton(n).click()
		}
		await expectRows(
			runs.slice((n - 1) * HISTORY_PAGE_SIZE, n * HISTORY_PAGE_SIZE),
		)
	}
	await expect(pageButton(pages + 1)).toHaveCount(0)
	if (pages > 1) {
		await pageButton(1).click()
	}
	const firstPage = runs.slice(0, HISTORY_PAGE_SIZE)
	await expectRows(firstPage)

	// Search matches the status and the start time.
	const search = page.getByPlaceholder("Search Tasks")
	const withStatus = (status: RunStatus) =>
		runs.filter(run => run.status === status)
	await search.fill("completed")
	await expectRows(withStatus("Completed").slice(0, HISTORY_PAGE_SIZE))
	await search.fill("failed")
	await expectRows(withStatus("Failed").slice(0, HISTORY_PAGE_SIZE))
	await search.fill("no-such-run")
	await expect(rows).toHaveCount(0)
	await search.fill("")
	await expectRows(firstPage)

	await page.getByTestId("refresh-tasks-button").click()
	await expectRows(firstPage)
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
