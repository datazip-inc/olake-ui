/**
 * Postgres → Apache Iceberg (JDBC catalog) integration workflow.
 *
 * Run by TestPostgres in server/tests, one tagged test at a time and in this order, between the
 * Go steps that seed and change the source and check the destination through Spark:
 *
 *   @create-job        CDC source, destination, and one job with a stream per feature
 *   @first-sync        sync 1                                  (Go checks, then changes the source)
 *   @second-sync       sync 2                                  (Go checks)
 *   @clear-destination Clear Destination from Job settings     (Go checks every table is gone)
 *   @rebuild-sync      sync 3 rebuilds the tables              (Go checks)
 *   @history           history, Job Type, search, refresh, task logs, errors filter, download
 *   @stream-rules      CDC modes off without CDC, Target Query Engines, "All Append",
 *                      Bulk Configure, index-build warning Cancel, wizard Cancel
 *   @alerts            webhook URL: save, edit + cancel, clear, save again
 *   @failed-sync       sync 4 with the source stopped by Go: Failed, "Failed jobs" tab
 *                                                             (Go waits for the webhook)
 *
 * Without E2E_JOB_NAME every test skips.
 */

import { TIMEOUTS } from "../../playwright.config"
import { DestinationConnector, SourceConnector } from "../enums"
import { testAuthenticated as test, expect } from "../fixtures"
import {
	E2E,
	STREAMS,
	SYNC_TIMEOUT,
	createIcebergJdbcDestination,
	createJobFromStreams,
	createPostgresSource,
	expectNewestRunLogs,
	expectWizardConnectionsPassed,
	historyRows,
	jobMenu,
	names,
	offeredUpsertTypes,
	openStream,
	pickOption,
	selectStream,
	setDataFilter,
	setIngestionMode,
	setSwitch,
	setSyncMode,
	setUpsertType,
	streamTab,
	switchNear,
	syncAndWait,
	syncModeRadio,
	waitForNewestRun,
} from "./helpers"

const job = E2E.job
const n = names(job)
const SYNC_COMPLETED_LOG =
	"Sync completed, wait 5 seconds cleanup in progress..."

test.describe("Postgres → Iceberg (JDBC) integration workflow", () => {
	// Go owns retries: a retry would create the job a second time.
	test.describe.configure({ retries: 0 })
	test.skip(
		!job,
		"run by the Go integration test (server/tests), which sets E2E_JOB_NAME",
	)

	// Every test starts on the jobs page, where the sidebar links the steps click are rendered.
	test.beforeEach(async ({ page }) => {
		await page.goto("/jobs")
		await expect(page).toHaveURL(/\/jobs$/)
	})

	test("@create-job: CDC source, destination, and a job with one stream per feature", async ({
		page,
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		jobsPage,
		createJobPage,
	}) => {
		test.setTimeout(20 * 60 * 1000)

		await createPostgresSource(
			page,
			sourcesPage,
			createSourcePage,
			n.source,
			true,
		)
		await createIcebergJdbcDestination(
			destinationsPage,
			createDestinationPage,
			n.destination,
		)

		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
		await jobsPage.clickCreateJob()
		await createJobPage.expectCreateJobPageVisible()
		await createJobPage.configureJobSettings(
			job,
			"Every Week",
			n.source,
			SourceConnector.Postgres,
			n.destination,
			DestinationConnector.ApacheIceberg,
		)
		await expectWizardConnectionsPassed(page)

		// Dynamic database name: the header shows the pattern, the modal has Dynamic selected,
		// and every stream's destination table is <job>_postgres_public/<table>.
		await expect(
			page.getByText(`${n.databasePrefix}_\${source_namespace}`),
		).toBeVisible({ timeout: TIMEOUTS.LONG })
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

		await page.getByRole("checkbox", { name: "Sync all" }).uncheck()

		const configure: Record<string, () => Promise<void>> = {
			// Full Refresh + CDC, Append: inserts, updates and deletes all land as rows.
			[STREAMS.cdc]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
			},
			// CDC Only, Append: sync 1 writes nothing, sync 2 only the new rows.
			[STREAMS.strictCdc]: async () => {
				await setSyncMode(page, "CDC Only")
				await setIngestionMode(page, "Append")
			},
			// Upsert with equality deletes, Iceberg v2.
			[STREAMS.upsertEq]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Upsert")
				await setUpsertType(page, "Equality Deletes", "V2")
			},
			// Upsert with deletion vectors, Iceberg v3.
			[STREAMS.upsertDv]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Upsert")
				await setUpsertType(page, "Deletion Vector", "V3")
			},
			// Data Filter amount > 20 AND id < 5 (needs Normalization, set below).
			[STREAMS.filter]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
				await setDataFilter(
					page,
					[
						{ column: "amount", operator: ">", value: "20" },
						{ column: "id", operator: "<", value: "5" },
					],
					"AND",
				)
			},
			// Sync new columns automatically on: a column added at the source reaches Iceberg.
			[STREAMS.newColumns]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
				await streamTab(page, "Schema")
				await setSwitch(
					switchNear(page, "Sync new columns automatically"),
					true,
				)
				await streamTab(page, "Config")
			},
		}

		for (const stream of Object.values(STREAMS)) {
			await selectStream(page, stream)
			await streamTab(page, "Config")
			await expect(
				page.getByText(`${n.database}/${stream}`, { exact: true }).first(),
			).toBeVisible()
			// Every stream is normalized: the checks read the source columns.
			await setSwitch(switchNear(page, "Normalization"), true)
			await configure[stream]()
		}

		// Deletion Vector needs a destination index, so Create Job asks first.
		await createJobFromStreams(page, true)
		await createJobPage.goToJobsPage()
		await jobsPage.expectJobsPageVisible()
		await jobsPage.expectJobExists(job)
	})

	test("@first-sync: sync 1", async ({ page, jobsPage }) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 1, "Completed")
		await expectNewestRunLogs(page, SYNC_COMPLETED_LOG)
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobStatus(job, "Completed")
	})

	test("@second-sync: sync 2, after the source changes", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 2, "Completed")
		await expectNewestRunLogs(page, SYNC_COMPLETED_LOG)
	})

	test("@clear-destination: Clear Destination from Job settings", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
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
		await waitForNewestRun(page, 3, "Clear Destination", "Completed")
	})

	test("@rebuild-sync: sync 3 rebuilds the cleared tables", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 4, "Completed")
		await expectNewestRunLogs(page, SYNC_COMPLETED_LOG)
	})

	test("@history: runs, Job Type, search, refresh and task logs", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(10 * 60 * 1000)
		await jobMenu(page, jobsPage, job, "Job Logs & History")

		for (const header of [
			"Start time",
			"Runtime",
			"Status",
			"Job Type",
			"Actions",
		]) {
			await expect(
				page.getByRole("columnheader", { name: header }),
			).toBeVisible()
		}
		const rows = historyRows(page)
		await expect(rows).toHaveCount(4)
		await expect(rows.filter({ hasText: "Completed" })).toHaveCount(4)
		await expect(rows.filter({ hasText: "Clear Destination" })).toHaveCount(1)
		await expect(rows.nth(1)).toContainText("Clear Destination")
		for (const index of [0, 2, 3]) {
			await expect(rows.nth(index)).toContainText("Sync")
		}

		// Search matches the status and the start time.
		const search = page.getByPlaceholder("Search Tasks")
		await search.fill("completed")
		await expect(rows).toHaveCount(4)
		await search.fill("no-such-run")
		await expect(rows).toHaveCount(0)
		await search.fill("")
		await expect(rows).toHaveCount(4)

		await page.getByTestId("refresh-tasks-button").click()
		await expect(rows).toHaveCount(4)

		// Task logs of the newest run
		await expectNewestRunLogs(page, SYNC_COMPLETED_LOG)
		const searchLogs = page.getByPlaceholder("Search Logs")
		await searchLogs.fill("Sync completed")
		await expect(page.getByText(SYNC_COMPLETED_LOG).first()).toBeVisible()
		await searchLogs.fill("")

		const errors = page.getByRole("button", { name: "Errors", exact: true })
		await errors.click()
		await expect(errors).toHaveClass(/ant-btn-primary/)
		await expect(page.getByText(SYNC_COMPLETED_LOG)).toHaveCount(0)
		await errors.click()
		await expect(errors).not.toHaveClass(/ant-btn-primary/)
		await expect(page.getByText(SYNC_COMPLETED_LOG).first()).toBeVisible()

		const download = page.waitForEvent("download")
		await page.getByRole("button", { name: "Download Logs" }).click()
		await expect(page.getByText("Downloading logs...")).toBeVisible()
		expect((await download).suggestedFilename()).not.toBe("")
	})

	test("@stream-rules: CDC modes, Target Query Engines, All Append, Bulk Configure", async ({
		page,
		sourcesPage,
		createSourcePage,
		jobsPage,
		createJobPage,
	}) => {
		test.setTimeout(25 * 60 * 1000)

		const back = async () => {
			await page.getByRole("button", { name: "Back", exact: true }).click()
			await page.getByRole("button", { name: "Yes, Leave" }).click()
			await expect(page.getByTestId("create-job-wizard-submit")).toHaveText(
				"Next",
			)
		}
		const pickSource = async (name: string) => {
			await pickOption(page, page.getByTestId("existing-source"), name)
		}
		// The only multi-select of the config step.
		const queryEngines = page.locator(".ant-select.ant-select-multiple").first()
		const toggleQueryEngine = async (label: string) => {
			await pickOption(page, queryEngines, label)
			await page.keyboard.press("Escape")
		}
		const next = async () => {
			await page.getByTestId("create-job-wizard-submit").click()
			await expectWizardConnectionsPassed(page)
			await expect(streamRow(STREAMS.upsertEq)).toBeVisible({
				timeout: TIMEOUTS.LONG,
			})
		}
		const streamRow = (stream: string) =>
			page.getByRole("button", { name: stream, exact: true })
		const upsertTypesOf = async (stream: string) => {
			await selectStream(page, stream)
			await streamTab(page, "Config")
			await setSyncMode(page, "Full Refresh + CDC")
			await setIngestionMode(page, "Upsert")
			return offeredUpsertTypes(page)
		}

		// A source with CDC off offers no CDC sync modes.
		await createPostgresSource(
			page,
			sourcesPage,
			createSourcePage,
			n.sourceNoCdc,
			false,
		)
		await jobsPage.navigateToJobs()
		await jobsPage.clickCreateJob()
		await createJobPage.expectCreateJobPageVisible()
		await createJobPage.configureJobSettings(
			n.rulesJob,
			"Every Week",
			n.sourceNoCdc,
			SourceConnector.Postgres,
			n.destination,
			DestinationConnector.ApacheIceberg,
		)
		await expectWizardConnectionsPassed(page)
		await selectStream(page, STREAMS.cdc)
		await streamTab(page, "Config")
		await expect(syncModeRadio(page, "Full Refresh + CDC")).toBeDisabled()
		await expect(syncModeRadio(page, "CDC Only")).toBeDisabled()
		await expect(syncModeRadio(page, "Full Refresh")).toBeEnabled()
		await expect(
			syncModeRadio(page, "Full Refresh + Incremental"),
		).toBeEnabled()

		// With the CDC source the CDC modes are offered.
		await back()
		await pickSource(n.source)
		await next()
		await selectStream(page, STREAMS.cdc)
		await streamTab(page, "Config")
		await expect(syncModeRadio(page, "Full Refresh + CDC")).toBeEnabled()
		await expect(syncModeRadio(page, "CDC Only")).toBeEnabled()

		// Target Query Engines decide the upsert types: none → all three, Dremio → Positional,
		// Databricks → Deletion Vector.
		expect(await upsertTypesOf(STREAMS.upsertEq)).toEqual([
			"Equality Deletes",
			"Positional Deletes",
			"Deletion Vector",
		])
		await back()
		await toggleQueryEngine("Dremio")
		await next()
		expect(await upsertTypesOf(STREAMS.upsertEq)).toEqual([
			"Positional Deletes",
		])
		await back()
		await toggleQueryEngine("Dremio")
		await toggleQueryEngine("Databricks")
		await next()
		expect(await upsertTypesOf(STREAMS.upsertEq)).toEqual(["Deletion Vector"])
		await back()
		await toggleQueryEngine("Databricks")
		await next()

		// "All Append" switches every stream to Append; "All Upsert" switches them back.
		await selectStream(page, STREAMS.cdc)
		await selectStream(page, STREAMS.upsertEq)
		const switchAll = async (mode: "Append" | "Upsert") => {
			await page.getByText(`All ${mode}`, { exact: true }).click()
			const dialog = page.getByRole("dialog").filter({
				hasText: `Switch to ${mode} for all tables ?`,
			})
			await expect(dialog).toBeVisible()
			await expect(dialog).toContainText(
				"You can change mode for individual tables",
			)
			await dialog.getByRole("button", { name: "Confirm" }).click()
			await expect(dialog).toBeHidden()
			for (const stream of [STREAMS.cdc, STREAMS.upsertEq]) {
				await openStream(page, stream)
				await streamTab(page, "Config")
				await expect(
					page.getByRole("radio", { name: mode, exact: true }),
				).toBeChecked()
			}
		}
		await switchAll("Append")
		await switchAll("Upsert")

		// Bulk Configure: Full Refresh for two streams.
		await page.getByRole("button", { name: "Bulk Configure" }).click()
		const bulk = page.getByRole("dialog").filter({ hasText: "Select Streams" })
		await expect(bulk).toBeVisible()
		await expect(bulk).toContainText(
			"Select streams you wish to bulk configure",
		)
		for (const stream of [STREAMS.cdc, STREAMS.filter]) {
			await bulk.getByRole("checkbox", { name: stream, exact: true }).check()
		}
		await bulk.getByRole("button", { name: "Configure Streams" }).click()
		const apply = page.getByRole("dialog").filter({
			hasText: "Apply Configurations",
		})
		await expect(apply).toBeVisible()
		await apply
			.getByRole("radio", { name: "Full Refresh", exact: true })
			.check()
		await apply.getByRole("button", { name: "Review Changes" }).click()
		const summary = page.getByRole("dialog").filter({
			hasText: "Summary of Changes",
		})
		await expect(summary).toBeVisible()
		await summary.getByRole("button", { name: "Apply Changes" }).click()
		await expect(
			page.getByText("2 streams configured successfully"),
		).toBeVisible()
		await expect(
			page.getByText("2 streams configured successfully"),
		).toBeHidden({
			timeout: 30_000,
		})
		for (const stream of [STREAMS.cdc, STREAMS.filter]) {
			await openStream(page, stream)
			await streamTab(page, "Config")
			await expect(syncModeRadio(page, "Full Refresh")).toBeChecked()
		}

		// The index-build warning: Cancel keeps the wizard open, nothing is created.
		await selectStream(page, STREAMS.upsertDv)
		await streamTab(page, "Config")
		await setSyncMode(page, "Full Refresh + CDC")
		await setIngestionMode(page, "Upsert")
		await setUpsertType(page, "Deletion Vector", "V3")
		await page.getByTestId("create-job-wizard-submit").click()
		const warning = page
			.getByRole("dialog")
			.filter({ hasText: "Please Confirm" })
		await expect(warning).toBeVisible()
		await warning.getByRole("button", { name: "Cancel" }).click()
		await expect(warning).toBeHidden()
		await expect(page.getByTestId("create-job-wizard-submit")).toHaveText(
			"Create Job",
		)

		// Cancel the wizard: the job is never created.
		await page
			.getByRole("button", { name: "Cancel", exact: true })
			.first()
			.click()
		const cancel = page.getByRole("dialog").filter({ hasText: "Don't cancel" })
		await expect(cancel).toBeVisible()
		await cancel.getByRole("button", { name: "Cancel", exact: true }).click()
		await expect(page).toHaveURL(/\/jobs$/)
		await expect(page.getByTestId(`job-${n.rulesJob}`)).toHaveCount(0)
	})

	test("@alerts: webhook URL save, edit and cancel, clear, save again", async ({
		page,
	}) => {
		test.skip(!E2E.webhookUrl, "needs E2E_WEBHOOK_URL")
		const url = E2E.webhookUrl
		const open = async () => {
			await page.goto("/settings")
			await page.getByRole("tab", { name: "Alerts and Notifications" }).click()
		}
		const input = page.getByPlaceholder("Enter your webhook URL")
		const save = page.getByRole("button", { name: "Save changes" })

		await open()
		// A new OLake has no URL, so the field starts editable and Save is disabled while empty.
		await expect(input).toBeEditable()
		await expect(input).toHaveValue("")
		await expect(save).toBeDisabled()
		await input.fill(url)
		await save.click()
		await expect(input).toBeDisabled()
		await expect(input).toHaveValue(url)
		await open()
		await expect(input).toHaveValue(url)
		await expect(input).toBeDisabled()

		// Edit, then Cancel, keeps the saved URL.
		await page.getByRole("button", { name: "Edit" }).click()
		await expect(input).toBeEditable()
		await input.fill("http://changed.invalid/alert")
		await page.getByRole("button", { name: "Cancel", exact: true }).click()
		await expect(input).toHaveValue(url)
		await expect(input).toBeDisabled()

		// Clear removes it.
		await page.getByRole("button", { name: "Clear webhook URL" }).click()
		await expect(input).toHaveValue("")
		await expect(input).toBeEditable()
		await open()
		await expect(input).toHaveValue("")

		// Save again: the URL the failure alert goes to.
		await input.fill(url)
		await save.click()
		await expect(input).toBeDisabled()
		await open()
		await expect(input).toHaveValue(url)
	})

	test("@failed-sync: a sync with the source down fails and shows in Failed jobs", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 5, "Failed")

		await jobsPage.navigateToJobs()
		await jobsPage.expectJobsPageVisible()
		await page.getByRole("tab", { name: "Failed jobs" }).click()
		await expect(page.getByTestId(`job-${job}`)).toBeVisible({
			timeout: TIMEOUTS.LONG,
		})
		await expect(
			page.locator(".ant-table-tbody tr").filter({ hasText: job }),
		).toContainText("Failed")
	})
})
