/** A job's settings page. */

import { Page } from "@playwright/test"

import { jobMenu, waitForNewestRun } from "./runs"
import { expect } from "../../fixtures"
import { JobsPage } from "../../pages/JobsPage"

/**
 * Clear Destination from the job's settings, confirmed in its dialog, then waits on the history
 * page until it is the newest of runs and Completed.
 */
export const clearDestination = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
	runs: number,
) => {
	await jobMenu(page, jobsPage, job, "Job settings")
	await page
		.getByRole("button", { name: "Clear Destination", exact: true })
		.click()
	const dialog = page.getByRole("dialog").filter({
		hasText: "This will erase all data that was synced by this job",
	})
	await expect(dialog).toBeVisible()
	await expect(dialog).toContainText("cannot be undone")
	await dialog.getByRole("button", { name: "Confirm" }).click()
	await expect(dialog).toBeHidden()

	await jobMenu(page, jobsPage, job, "Job Logs & History")
	await waitForNewestRun(page, runs, "Clear Destination", "Completed")
}
