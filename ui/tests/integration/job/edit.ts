/** Edit Streams on an existing job: the job edit page, which opens on the streams step. */

import { Page } from "@playwright/test"

import { jobMenu } from "./runs"
import { streamRow } from "./streams"
import { expect } from "../../fixtures"
import { JobsPage } from "../../pages/JobsPage"
import { TIMEOUTS } from "../e2e"

/** The edit page's main button: "Next" on the config step, "Finish" on the streams step. */
const mainButton = (page: Page, label: "Next" | "Finish") =>
	page.getByRole("button", { name: label, exact: true })

/** Opens Edit Streams for job and waits until discovery lists stream. */
export const openEditStreams = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
	stream: string,
) => {
	await jobMenu(page, jobsPage, job, "Edit Streams")
	await expect(page).toHaveURL(/\/jobs\/\d+\/edit/)
	await expect(streamRow(page, stream)).toBeVisible({ timeout: TIMEOUTS.LONG })
}

/** "Back" from the streams step to the job config step, leaving the streams step's changes. */
export const editBack = async (page: Page) => {
	await page.getByRole("button", { name: "Back", exact: true }).click()
	await page.getByRole("button", { name: "Yes, Leave" }).click()
	await expect(mainButton(page, "Next")).toBeVisible()
}

/**
 * "Next" after the Target Query Engines changed: the warning says the delete formats may change,
 * and Proceed discovers the streams again, so the streams step lists stream once more.
 */
export const proceedWithChangedQueryEngines = async (
	page: Page,
	stream: string,
) => {
	await mainButton(page, "Next").click()
	const warning = page.getByRole("dialog").filter({
		hasText: "Target query engines changed",
	})
	await expect(warning).toBeVisible()
	await expect(warning).toContainText(
		"may change the resolved delete format for your streams",
	)
	await warning.getByRole("button", { name: "Proceed" }).click()
	await expect(warning).toBeHidden()
	await expect(streamRow(page, stream)).toBeVisible({ timeout: TIMEOUTS.LONG })
}

/**
 * "Finish" after stream's configuration changed: the dialog names stream as the one whose
 * destination data is cleared, and "Confirm and Finish" saves the job.
 */
export const finishWithStreamChange = async (page: Page, stream: string) => {
	await mainButton(page, "Finish").click()
	const dialog = page.getByRole("dialog").filter({
		hasText: "Are you sure you want to continue?",
	})
	await expect(dialog).toBeVisible()
	await expect(dialog).toContainText(
		"Modifying stream configurations will clear destination data for the impacted streams",
	)
	await expect(dialog).toContainText(stream)
	await expect(dialog).toContainText("Any ongoing sync will be auto cancelled.")
	await dialog.getByRole("button", { name: "Confirm and Finish" }).click()
	await expect(page).toHaveURL(/\/jobs$/)
}

/** "Finish" with no stream's configuration changed: the job is saved without asking. */
export const finishEdit = async (page: Page) => {
	await mainButton(page, "Finish").click()
	await expect(page).toHaveURL(/\/jobs$/)
}

/**
 * With the source unreachable, Edit Streams cannot discover the streams: the dialog says so and
 * Retry fails again. Close leaves the empty streams list, with a Retry button of its own.
 */
export const expectDiscoveryFailure = async (
	page: Page,
	jobsPage: JobsPage,
	job: string,
) => {
	await jobMenu(page, jobsPage, job, "Edit Streams")
	const dialog = page.getByRole("dialog").filter({
		hasText: "Failed to discover streams",
	})
	await expect(dialog).toBeVisible({ timeout: TIMEOUTS.LONG })
	await expect(dialog).toContainText(
		"Please check your connection and try again",
	)
	await dialog.getByRole("button", { name: "Retry" }).click()
	await expect(dialog).toBeHidden()
	await expect(dialog).toBeVisible({ timeout: TIMEOUTS.LONG })
	await dialog.getByRole("button", { name: "Close" }).click()
	await expect(dialog).toBeHidden()
	await expect(page.getByRole("button", { name: "Retry" })).toBeVisible()
}
