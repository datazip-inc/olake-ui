/** A run's logs page. */

import { Page } from "@playwright/test"

import { historyRows } from "./runs"
import { expect } from "../../fixtures"

/** The line a finished sync writes to its log. */
export const SYNC_COMPLETED_LOG =
	"Sync completed, wait 5 seconds cleanup in progress..."

/**
 * Opens the newest run's logs and checks that they hold line, the line a finished sync writes.
 *
 * The page loads the newest 1000 log lines into a virtual list, which renders only the rows
 * around the scroll position; with a long log the last line can be loaded but never rendered.
 * So search for the line: "Search Logs" filters the loaded lines, and the match is the only row.
 */
export const expectNewestRunLogs = async (
	page: Page,
	line = SYNC_COMPLETED_LOG,
) => {
	await historyRows(page)
		.first()
		.getByRole("button", { name: "View logs" })
		.click()
	await expect(page).toHaveURL(/\/jobs\/\d+\/history\/\d+\/logs\?file=/)
	const search = page.getByPlaceholder("Search Logs")
	await search.fill(line)
	try {
		await expect(async () => {
			await page.getByTestId("refresh-logs-button").click()
			await expect(page.getByText(line).first()).toBeVisible({ timeout: 3000 })
		}).toPass({ timeout: 2 * 60 * 1000, intervals: [2000] })
	} catch (error) {
		// Show what the log holds instead: its newest rendered lines, without the search. Clearing
		// the search reloads the log and then scrolls it to the end, so give that a moment.
		await search.fill("")
		await page.waitForTimeout(2000)
		const rendered = await page
			.locator("[data-testid='virtuoso-item-list']")
			.innerText({ timeout: 5000 })
			.catch(() => "(no log lines rendered)")
		throw new Error(
			`the run's log has no line "${line}"; rendered lines:\n${rendered.slice(-3000)}\n${error}`,
		)
	}
	await search.fill("")
}

/**
 * On the logs page of a finished sync: Search Logs finds the completion line, the Errors filter
 * hides it and shows it again, and Download Logs downloads the log.
 *
 * The checks search the log: it is a virtual list that renders only the rows in view, so a line
 * of a long log may be loaded but not rendered.
 */
export const expectLogTools = async (page: Page) => {
	const searchLogs = page.getByPlaceholder("Search Logs")
	await searchLogs.fill("Sync completed")
	await expect(page.getByText(SYNC_COMPLETED_LOG).first()).toBeVisible()

	// Errors also keeps only error lines, so the completion line goes away, and comes back.
	const errors = page.getByRole("button", { name: "Errors", exact: true })
	await errors.click()
	await expect(errors).toHaveClass(/ant-btn-primary/)
	await expect(page.getByText(SYNC_COMPLETED_LOG)).toHaveCount(0)
	await expect(page.getByText("No logs found")).toBeVisible()
	await errors.click()
	await expect(errors).not.toHaveClass(/ant-btn-primary/)
	await expect(page.getByText(SYNC_COMPLETED_LOG).first()).toBeVisible()
	await searchLogs.fill("")

	const download = page.waitForEvent("download")
	await page.getByRole("button", { name: "Download Logs" }).click()
	await expect(page.getByText("Downloading logs...")).toBeVisible()
	expect((await download).suggestedFilename()).not.toBe("")
}
