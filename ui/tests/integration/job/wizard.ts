/** The job wizard: the config step, moving between steps, and creating or cancelling the job. */

import { Page } from "@playwright/test"

import { streamRow } from "./streams"
import { DestinationConnector, SourceConnector } from "../../enums"
import { expect } from "../../fixtures"
import { CreateJobPage } from "../../pages/CreateJobPage"
import { JobsPage } from "../../pages/JobsPage"
import {
	assertTestConnectionOutcome,
	expectTestConnectionModalVisible,
} from "../../utils"
import { pickOption } from "../common/controls"
import { TIMEOUTS } from "../e2e"
import { expectSourceTestPassed } from "../sources/source"

/** What the config step of a new job is filled with. */
export type JobConfig = {
	job: string
	source: string
	sourceConnector: SourceConnector
	destination: string
	destinationConnector: DestinationConnector
	frequency?: string // "Every Week" when unset
}

/** The wizard's main button: "Next" on the config step, "Create Job" on the streams step. */
const submit = (page: Page) => page.getByTestId("create-job-wizard-submit")

/** Waits for the job wizard's source and destination test connections after "Next". */
export const expectWizardConnectionsPassed = async (page: Page) => {
	await expectTestConnectionModalVisible(page, "Source")
	await expectSourceTestPassed(page)
	await expectTestConnectionModalVisible(page, "Destination")
	await assertTestConnectionOutcome(page, "Destination")
}

/**
 * Opens Create Job from the jobs page and fills the config step with existing source and
 * destination; "Next" tests both connections, which must pass, and opens the streams step.
 */
export const startJob = async (
	page: Page,
	jobsPage: JobsPage,
	createJobPage: CreateJobPage,
	config: JobConfig,
) => {
	await jobsPage.navigateToJobs()
	await jobsPage.expectJobsPageVisible()
	await jobsPage.clickCreateJob()
	await createJobPage.expectCreateJobPageVisible()
	await createJobPage.configureJobSettings(
		config.job,
		config.frequency ?? "Every Week",
		config.source,
		config.sourceConnector,
		config.destination,
		config.destinationConnector,
	)
	await expectWizardConnectionsPassed(page)
}

/** "Back" from the streams step to the config step, leaving the streams step's changes. */
export const wizardBack = async (page: Page) => {
	await page.getByRole("button", { name: "Back", exact: true }).click()
	await page.getByRole("button", { name: "Yes, Leave" }).click()
	await expect(submit(page)).toHaveText("Next")
}

/** "Next" from the config step: the connections must pass and the streams list show stream. */
export const wizardNext = async (page: Page, stream: string) => {
	await submit(page).click()
	await expectWizardConnectionsPassed(page)
	await expect(streamRow(page, stream)).toBeVisible({ timeout: TIMEOUTS.LONG })
}

/** Picks an existing source on the config step. */
export const pickExistingSource = async (page: Page, name: string) => {
	await pickOption(page, page.getByTestId("existing-source"), name)
}

/** Turns a Target Query Engine on or off on the config step; it is the step's only multi-select. */
export const toggleQueryEngine = async (page: Page, label: string) => {
	await pickOption(
		page,
		page.locator(".ant-select.ant-select-multiple").first(),
		label,
	)
	await page.keyboard.press("Escape")
}

/**
 * On the streams step of a job to an Iceberg destination, checks the Dynamic database name: the
 * header shows <prefix>_${source_namespace}, and the database modal has Dynamic selected.
 */
export const expectDynamicDatabase = async (page: Page, prefix: string) => {
	await expect(page.getByText(`${prefix}_\${source_namespace}`)).toBeVisible({
		timeout: TIMEOUTS.LONG,
	})
	await page
		.getByText("Iceberg DB", { exact: true })
		.locator("xpath=ancestor::div[contains(@class,'group')][1]")
		.locator("svg.cursor-pointer")
		.first()
		.click()
	const databaseModal = page.getByRole("dialog").filter({
		hasText: "Iceberg Database Name",
	})
	await expect(databaseModal).toBeVisible()
	await expect(
		databaseModal.getByRole("radio", { name: /Dynamic \(Default\)/ }),
	).toBeChecked()
	await databaseModal.getByRole("button", { name: "Close" }).click()
	await expect(databaseModal).toBeHidden()
}

/**
 * "Create Job" on the streams step, confirming the index-build warning when it is expected, then
 * checks the jobs page lists job.
 */
export const createJobFromStreams = async (
	page: Page,
	createJobPage: CreateJobPage,
	jobsPage: JobsPage,
	job: string,
	expectIndexBuildWarning: boolean,
) => {
	await submit(page).click()
	if (expectIndexBuildWarning) {
		const dialog = page.getByRole("dialog").filter({
			hasText: "Please Confirm",
		})
		await expect(dialog).toBeVisible()
		await expect(dialog).toContainText(
			"Indexes might get built for streams using position deletes or deletion vectors",
		)
		await dialog.getByRole("button", { name: "Confirm" }).click()
	}
	await createJobPage.goToJobsPage()
	await jobsPage.expectJobsPageVisible()
	await jobsPage.expectJobExists(job)
}

/** "Create Job" with the index-build warning expected, then Cancel: the wizard stays open. */
export const cancelAtIndexBuildWarning = async (page: Page) => {
	await submit(page).click()
	const warning = page.getByRole("dialog").filter({ hasText: "Please Confirm" })
	await expect(warning).toBeVisible()
	await warning.getByRole("button", { name: "Cancel" }).click()
	await expect(warning).toBeHidden()
	await expect(submit(page)).toHaveText("Create Job")
}

/** Cancels the wizard: back on the jobs page, which does not list job. */
export const cancelWizard = async (page: Page, job: string) => {
	await page
		.getByRole("button", { name: "Cancel", exact: true })
		.first()
		.click()
	const cancel = page.getByRole("dialog").filter({ hasText: "Don't cancel" })
	await expect(cancel).toBeVisible()
	await cancel.getByRole("button", { name: "Cancel", exact: true }).click()
	await expect(page).toHaveURL(/\/jobs$/)
	await expect(page.getByTestId(`job-${job}`)).toHaveCount(0)
}
