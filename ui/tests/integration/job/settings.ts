/** A job's settings page. */

import { Page } from "@playwright/test"

import { jobMenu, waitForNewestRun } from "./runs"
import { setFrequency } from "./wizard"
import { expect } from "../../fixtures"
import { JobsPage } from "../../pages/JobsPage"

/**
 * Clear Destination from the job's settings: Cancel in its dialog first, which clears nothing,
 * then Confirm. While it runs, editing the job is disabled. Then waits on the history page until
 * it is the newest of runs and Completed.
 */
export const clearDestination = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
	runs: number,
) => {
	await jobMenu(page, jobsPage, job, "Job settings")
	await expect(page).toHaveURL(/\/jobs\/\d+\/settings/)
	const jobId = page.url().match(/\/jobs\/(\d+)\//)![1]
	const clear = page.getByRole("button", {
		name: "Clear Destination",
		exact: true,
	})
	const dialog = page.getByRole("dialog").filter({
		hasText: "This will erase all data that was synced by this job",
	})
	await clear.click()
	await expect(dialog).toBeVisible()
	await expect(dialog).toContainText("cannot be undone")
	await dialog.getByRole("button", { name: "Cancel", exact: true }).click()
	await expect(dialog).toBeHidden()

	// Wait for the request: leaving the page before it answers aborts it, and the clear with it.
	await clear.click()
	await expect(dialog).toBeVisible()
	const started = page.waitForResponse(
		response =>
			response.url().includes("/clear-destination") &&
			response.request().method() === "POST",
	)
	await dialog.getByRole("button", { name: "Confirm" }).click()
	expect((await started).ok(), "Clear Destination started").toBe(true)
	await expect(dialog).toBeHidden()

	await expectEditingDisabled(page, jobId)

	await jobMenu(page, jobsPage, job, "Job Logs & History")
	await waitForNewestRun(page, runs, "Clear Destination", "Completed")
}

/**
 * While Clear Destination runs, Job settings and Edit Streams open the "Editing Disabled" dialog
 * instead, whose "Go back to Jobs" returns to the jobs page. A clear takes seconds only, so both
 * pages are opened straight from their URLs.
 */
export const expectEditingDisabled = async (page: Page, jobId: string) => {
	const dialog = page.getByRole("dialog").filter({
		hasText: "Editing Disabled",
	})
	for (const [path, what] of [
		["settings", "Job Settings Edit"],
		["edit", "Stream editing"],
	] as const) {
		await page.goto(`/jobs/${jobId}/${path}`)
		await expect(dialog).toBeVisible()
		await expect(dialog).toContainText(
			`${what} is disabled while the destination is being cleared`,
		)
		await dialog.getByRole("button", { name: "Go back to Jobs" }).click()
		await expect(page).toHaveURL(/\/jobs$/)
	}
}

/** In the job's settings, checks the frequency is from, then sets it to label and saves it. */
export const setJobFrequency = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
	from: string,
	label: string,
) => {
	await jobMenu(page, jobsPage, job, "Job settings")
	await expect(
		page
			.getByText("Frequency", { exact: true })
			.locator("xpath=ancestor::div[.//div[contains(@class,'ant-select')]][1]"),
	).toContainText(from)
	await setFrequency(page, label)
	await page.getByRole("button", { name: "Save", exact: true }).click()
	await expect(page).toHaveURL(/\/jobs$/)
}

/** In the job's settings, Clear Destination is disabled: the job is paused. */
export const expectClearDestinationDisabled = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
) => {
	await jobMenu(page, jobsPage, job, "Job settings")
	await expect(
		page.getByRole("button", { name: "Clear Destination", exact: true }),
	).toBeDisabled()
}
