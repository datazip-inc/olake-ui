/**
 * UI actions for the integration workflows in this folder.
 *
 * The Go integration test (server/tests) starts the product, then runs one tagged test of a spec
 * here at a time, with the inputs in E2E_* variables. Selectors follow the components under
 * ui/src; field ids follow RJSF (#root_<field>, nested #root_<parent>_<field>).
 */

import { Locator, Page } from "@playwright/test"

import { TIMEOUTS } from "../../playwright.config"
import { expect } from "../fixtures"
import { CreateDestinationPage } from "../pages/CreateDestinationPage"
import { CreateSourcePage } from "../pages/CreateSourcePage"
import { DestinationsPage } from "../pages/DestinationsPage"
import { JobsPage } from "../pages/JobsPage"
import { SourcesPage } from "../pages/SourcesPage"
import {
	assertTestConnectionOutcome,
	createIcebergJdbcConfig,
	createPostgresSourceConfig,
	expectTestConnectionModalVisible,
	ICEBERG_JDBC_CONFIG,
	POSTGRES_SOURCE_CONFIG,
	verifyEntityCreationSuccessModal,
} from "../utils"

/**
 * Inputs from the Go integration test (server/tests, Run.Playwright), which runs Playwright as a
 * separate process and passes them as environment variables. A spec skips itself when
 * E2E_JOB_NAME is unset.
 */
export const E2E = {
	job: process.env.E2E_JOB_NAME?.trim() || "",
	// The destination database the Go checks read: <job>_<source database>_<schema>.
	database: process.env.E2E_DATABASE?.trim() || "",
	webhookUrl: process.env.E2E_WEBHOOK_URL?.trim() || "",
}

/** Names of everything the Postgres workflow creates, derived from the job name. */
export const names = (job: string, database: string) => ({
	source: `${job}_source`,
	// Unique per run: the source is created in @stream-rules, which can be run again on its own.
	sourceNoCdc: `${job}_source_nocdc_${Date.now()}`,
	destination: `${job}_dest`,
	rulesJob: `${job}_rules`,
	database,
	// The UI shows a Dynamic name as <prefix>_${source_namespace}: the database without its
	// source schema (public).
	databasePrefix: database.replace(/_public$/, ""),
})

/** The e2e tables the Go test seeds, one stream per feature. */
export const STREAMS = {
	cdc: "e2e_cdc",
	strictCdc: "e2e_strict_cdc",
	upsertEq: "e2e_upsert_eq",
	upsertDv: "e2e_upsert_dv",
	filter: "e2e_filter",
	newColumns: "e2e_new_columns",
} as const

export const CDC = {
	slot: "olake_slot",
	publication: "olake_publication",
	initialWaitTime: "120", // the minimum; lower values silently become 10800
}

/** Long waits: a CDC sync waits initial_wait_time for changes after its full load. */
export const SYNC_TIMEOUT = 15 * 60 * 1000

// ---------------------------------------------------------------------------------------------
// Sources and destinations

/** Turns the source form's CDC on or off: the "Enable CDC" switch, or the older radios. */
export const setSourceCdc = async (page: Page, on: boolean) => {
	const cdcSwitch = page.getByRole("switch", { name: "Enable CDC" })
	if (await cdcSwitch.isVisible()) {
		if ((await cdcSwitch.getAttribute("aria-checked")) !== String(on)) {
			await cdcSwitch.click()
		}
		await expect(cdcSwitch).toHaveAttribute("aria-checked", String(on))
		return
	}
	await page.getByRole("radio", { name: on ? "CDC" : "Standalone" }).check()
}

/**
 * Waits for the source test connection to finish and fails unless it succeeded. With CDC on the
 * prerequisite checks must all pass: "Continue anyway" means one did not.
 */
export const expectSourceTestPassed = async (page: Page) => {
	const wait = (locator: Locator, outcome: string) =>
		locator
			.waitFor({ state: "visible", timeout: TIMEOUTS.LONG })
			.then(() => outcome)
			.catch(() => "timeout")
	const outcome = await Promise.race([
		wait(page.getByText("Source test connection is successful"), "success"),
		wait(page.getByText("Source test connection has failed"), "failed"),
		wait(
			page.getByRole("button", { name: "Continue anyway" }),
			"checks failed",
		),
	])
	// On failure, name the checks: the prerequisite dialog lists each one with its current value.
	const details =
		outcome === "success"
			? ""
			: await page
					.getByRole("dialog", { name: "Testing Source Configuration" })
					.innerText({ timeout: 2000 })
					.catch(() => "")
	expect(
		outcome,
		`source test connection and its CDC prerequisite checks must all pass\n${details}`,
	).toBe("success")
}

/** Creates the Postgres source the Go test started, with CDC on or off. */
export const createPostgresSource = async (
	page: Page,
	sourcesPage: SourcesPage,
	createSourcePage: CreateSourcePage,
	name: string,
	cdc: boolean,
) => {
	await sourcesPage.navigateToSources()
	await sourcesPage.expectSourcesPageVisible()
	await sourcesPage.clickCreateSource()
	await createSourcePage.expectCreateSourcePageVisible()
	await createSourcePage.fillSourceForm(
		createPostgresSourceConfig({ name, ...POSTGRES_SOURCE_CONFIG }),
	)
	await setSourceCdc(page, cdc)
	if (cdc) {
		await page.locator("#root_update_method_replication_slot").fill(CDC.slot)
		await page.locator("#root_update_method_publication").fill(CDC.publication)
		await page
			.locator("#root_update_method_initial_wait_time")
			.fill(CDC.initialWaitTime)
	}
	await createSourcePage.clickCreate()
	await expectTestConnectionModalVisible(page, "Source")
	await expectSourceTestPassed(page)
	await createSourcePage.expectEntitySavedModal()
	await sourcesPage.expectSourcesPageVisible()
	await sourcesPage.expectSourceExists(name)
}

/** Creates the Iceberg (JDBC catalog) destination the Go test started. */
export const createIcebergJdbcDestination = async (
	destinationsPage: DestinationsPage,
	createDestinationPage: CreateDestinationPage,
	name: string,
) => {
	await destinationsPage.navigateToDestinations()
	await destinationsPage.expectDestinationsPageVisible()
	await destinationsPage.clickCreateDestination()
	await createDestinationPage.expectCreateDestinationPageVisible()
	await createDestinationPage.fillDestinationForm(
		createIcebergJdbcConfig({ name, ...ICEBERG_JDBC_CONFIG }),
	)
	await createDestinationPage.clickCreate()
	await verifyEntityCreationSuccessModal(createDestinationPage)
	await destinationsPage.expectDestinationsPageVisible()
	await destinationsPage.expectDestinationExists(name)
}

/** Waits for the job wizard's source and destination test connections after "Next". */
export const expectWizardConnectionsPassed = async (page: Page) => {
	await expectTestConnectionModalVisible(page, "Source")
	await expectSourceTestPassed(page)
	await expectTestConnectionModalVisible(page, "Destination")
	await assertTestConnectionOutcome(page, "Destination")
}

// ---------------------------------------------------------------------------------------------
// Job wizard: streams step

/** The row of a stream in the streams list; clicking it opens the stream's configuration. */
export const streamRow = (page: Page, stream: string) =>
	page.getByRole("button", { name: stream, exact: true })

/** Ticks a stream and opens its configuration. */
export const selectStream = async (page: Page, stream: string) => {
	const row = streamRow(page, stream)
	await row.getByRole("checkbox").check()
	await row.click()
	await expect(
		page.getByText(new RegExp(`/${stream}$`)).first(),
		`${stream} configuration is open`,
	).toBeVisible()
}

/** Opens a stream's configuration without changing its selection. */
export const openStream = async (page: Page, stream: string) => {
	await streamRow(page, stream).click()
	await expect(page.getByText(new RegExp(`/${stream}$`)).first()).toBeVisible()
}

/** A tab of the open stream's configuration. */
export const streamTab = async (
	page: Page,
	tab: "Config" | "Schema" | "Partitioning",
) => {
	await page.getByRole("button", { name: tab, exact: true }).click()
}

/** The switch in the closest section around a label. */
export const switchNear = (page: Page, label: string) =>
	page
		.getByText(label, { exact: true })
		.first()
		.locator("xpath=ancestor::div[.//button[@role='switch']][1]")
		.getByRole("switch")
		.first()

/** Sets a switch to on or off and checks the result. */
export const setSwitch = async (control: Locator, on: boolean) => {
	if ((await control.getAttribute("aria-checked")) !== String(on)) {
		await control.click()
	}
	await expect(control).toHaveAttribute("aria-checked", String(on))
}

export type SyncModeLabel =
	| "Full Refresh"
	| "Full Refresh + Incremental"
	| "Full Refresh + CDC"
	| "CDC Only"

export const syncModeRadio = (page: Page, mode: SyncModeLabel) =>
	page.getByRole("radio", { name: mode, exact: true })

export const setSyncMode = async (page: Page, mode: SyncModeLabel) => {
	await syncModeRadio(page, mode).check()
	await expect(syncModeRadio(page, mode)).toBeChecked()
}

export const setIngestionMode = async (
	page: Page,
	mode: "Upsert" | "Append",
) => {
	const radio = page.getByRole("radio", { name: mode, exact: true })
	await radio.check()
	await expect(radio).toBeChecked()
}

export type UpsertTypeLabel =
	| "Equality Deletes"
	| "Positional Deletes"
	| "Deletion Vector"

/** The "Upsert Type:" select of the open stream. */
const upsertTypeSelect = (page: Page) =>
	page
		.getByText("Upsert Type:", { exact: true })
		.locator("xpath=ancestor::div[.//div[contains(@class,'ant-select')]][1]")
		.locator(".ant-select")
		.first()

/** The upsert types the open stream offers (at most 3, so all of them are rendered). */
export const offeredUpsertTypes = async (page: Page): Promise<string[]> => {
	const select = upsertTypeSelect(page)
	await select.click()
	const options = (await dropdownOf(page, select)).locator(
		".ant-select-item-option",
	)
	await expect(options.first()).toBeVisible()
	const labels = (await options.allInnerTexts()).map(text => text.trim())
	await page.keyboard.press("Escape")
	return labels
}

export const setUpsertType = async (
	page: Page,
	type: UpsertTypeLabel,
	formatVersion: "V2" | "V3",
) => {
	await pickOption(page, upsertTypeSelect(page), type)
	await expect(upsertTypeSelect(page)).toContainText(type)
	await expect(page.getByText("Iceberg format version:")).toContainText(
		formatVersion,
	)
}

/**
 * The dropdown of an antd Select. antd keeps every dropdown it has opened in the page, closed
 * ones included, so find this select's own through the list id its input controls.
 */
export const dropdownOf = async (page: Page, select: Locator) => {
	const id = await select
		.locator("input[role='combobox']")
		.first()
		.getAttribute("aria-controls")
	expect(id, "the select's input names its option list").toBeTruthy()
	return page
		.locator(".ant-select-dropdown")
		.filter({ has: page.locator(`[id="${id}"]`) })
}

/**
 * Opens select and clicks its option with exactly text.
 * - antd also renders a hidden copy of each option (role="option", for screen readers), so this
 *   matches the rendered .ant-select-item-option instead.
 * - The list is virtual: only the rows in view exist (about 8), and it scrolls on wheel events
 *   only, not on scrollTop. So scroll it to the top, then down until the option is rendered.
 */
export const pickOption = async (page: Page, select: Locator, text: string) => {
	await select.click()
	const dropdown = await dropdownOf(page, select)
	const options = dropdown.locator(".ant-select-item-option")
	await expect(options.first()).toBeVisible()
	const option = options
		.filter({ has: page.getByText(text, { exact: true }) })
		.first()
	// Wheel over the list, aiming at it again each time (a wheel the list cannot use scrolls the
	// page instead, which moves the list). Turn around at either end: an earlier pick can leave
	// the list scrolled down.
	const list = dropdown.locator(".rc-virtual-list-holder").first()
	let direction = 1
	let rendered = ""
	for (let i = 0; i < 100 && !(await option.isVisible()); i++) {
		const box = await list.boundingBox()
		if (!box) break
		await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
		await page.mouse.wheel(0, direction * 80)
		await page.waitForTimeout(120)
		const now = (await options.allInnerTexts()).join("|")
		if (now === rendered) direction = -direction
		rendered = now
	}
	await option.click()
}

export type FilterCondition = {
	column: string
	operator: ">" | "<" | "=" | "!=" | ">=" | "<="
	value: string
}

/**
 * Turns the open stream's Data Filter on with up to two conditions joined by AND or OR. Data
 * Filter needs Normalization on.
 */
export const setDataFilter = async (
	page: Page,
	conditions: FilterCondition[],
	join: "AND" | "OR",
) => {
	await setSwitch(switchNear(page, "Data Filter"), true)
	for (const [index, condition] of conditions.entries()) {
		if (index > 0) {
			await page.getByRole("button", { name: "New Column filter" }).click()
		}
		const block = page
			.getByText(index === 0 ? "Column I" : "Column II", { exact: true })
			.locator("xpath=ancestor::div[.//label[normalize-space()='Value']][1]")
		const selects = block.locator(".ant-select")
		await pickOption(page, selects.nth(0), condition.column)
		await pickOption(page, selects.nth(1), condition.operator)
		await block.getByRole("textbox").fill(condition.value)
		await expect(selects.nth(0)).toContainText(condition.column)
		await expect(selects.nth(1)).toContainText(condition.operator)
		await expect(block.getByRole("textbox")).toHaveValue(condition.value)
	}
	if (conditions.length > 1) {
		await page.getByRole("button", { name: join, exact: true }).click()
	}
}

/** "Create Job" on the streams step; confirms the index-build warning when it is expected. */
export const createJobFromStreams = async (
	page: Page,
	expectIndexBuildWarning: boolean,
) => {
	await page.getByTestId("create-job-wizard-submit").click()
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
}

// ---------------------------------------------------------------------------------------------
// Jobs, history and logs

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
	jobType: "Sync" | "Clear Destination",
	status: "Completed" | "Failed",
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
	status: "Completed" | "Failed",
) => {
	await jobsPage.navigateToJobs()
	await jobsPage.expectJobsPageVisible()
	await jobsPage.syncJob(job)
	await waitForNewestRun(page, runs, "Sync", status)
}

/** Opens the newest run's logs and waits for the line a finished sync writes. */
export const expectNewestRunLogs = async (page: Page, line: string) => {
	await historyRows(page)
		.first()
		.getByRole("button", { name: "View logs" })
		.click()
	await expect(page).toHaveURL(/\/jobs\/\d+\/history\/\d+\/logs\?file=/)
	await expect(async () => {
		await page.getByTestId("refresh-logs-button").click()
		await expect(page.getByText(line).first()).toBeVisible({ timeout: 2000 })
	}).toPass({ timeout: TIMEOUTS.LONG, intervals: [3000] })
}
